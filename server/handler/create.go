package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
)

// Create creates a new app, not yet built or deployed
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
	"deploy": "true",
*/
func Create(w http.ResponseWriter, r *http.Request) {
	tsk := r.FormValue("app")
	if len(tsk) == 0 {
		http.Error(w, "Require a app definition", http.StatusBadRequest)
		return
	}

	var aapp *app.App
	err := json.Unmarshal([]byte(tsk), &aapp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = app.Create(aapp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events.Send(aapp.Id, events.Event{Body: "App created successfully", Type: events.Message})

	deploy, err := strconv.ParseBool(r.FormValue("deploy"))
	if err != nil {
		return
	}

	if deploy {
		go build(aapp, deploy)
	}
}
