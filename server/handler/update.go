package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
	log "github.com/cihub/seelog"
)

// Update modifies a app.
/*
	"app": {
		"id": "foo",
		"description": "This is some app",
		"source": {
			"code": {
				"lang": "golang",
				"text": "package main...",
			},
			"Dockerfile": "FROM ubuntu",
			"GitUrl": "https://github.com/foo/bar.git",
		}
	}
*/
func update(a *app.App, deploy bool) {
	err := app.Update(a)
	if err != nil {
		log.Errorf("Error updating app: %v", err)
		return
	}

	events.Send(a.Id, events.Event{Body: "App updated successfully", Type: events.Message})

	if !deploy {
		return
	}

	log.Infof("Building code for id: %s", a.Id)
	if err = a.Build(); err != nil {
		log.Errorf("Error building image: %v", err)
		events.Send(a.Id, events.Event{Body: "An error occurred during the build: " + err.Error(), Type: events.Error})
		return
	}

	events.Send(a.Id, events.Event{Body: "Build complete", Type: events.Message})

	// push to registry
	if err := a.Push(); err != nil {
		log.Errorf("Error pushing image: %v", err)
		events.Send(a.Id, events.Event{Body: "An error occurred during the push process: " + err.Error(), Type: events.Error})
		return
	}

	events.Send(a.Id, events.Event{Body: "Push complete", Type: events.Message})

	if err = a.Restart(); err != nil {
		events.Send(a.Id, events.Event{Body: err.Error(), Type: events.Error})
		return
	}

	events.Send(a.Id, events.Event{Body: "Deploy complete", Type: events.Message})
}

func Update(w http.ResponseWriter, r *http.Request) {
	// TODO: implement update
	tsk := r.FormValue("app")
	if len(tsk) == 0 {
		http.Error(w, "Require a app definition", http.StatusBadRequest)
		return
	}

	var a *app.App
	err := json.Unmarshal([]byte(tsk), &a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deploy, err := strconv.ParseBool(r.FormValue("deploy"))
	if err != nil {
		deploy = false
	}

	go update(a, deploy)
}
