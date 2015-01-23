package handler

import (
	"net/http"
	"strconv"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
	log "github.com/cihub/seelog"
)

// Build create a docker image for a app
/*
	"id": "foo",

	OR

	"app": {
		"id": "foo",
		"description": "Some thing",
		"source": {
			"code": {
				"lang": "ruby",
				"text": "puts 'hello world'",
			}
		}
	}
	"deploy": true

*/
func build(a *app.App, deploy bool) {
	// build app
	log.Infof("Building code for id: %s", a.Id)
	if err := a.Build(); err != nil {
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

	if !deploy {
		return
	}

	// deploy to kubernetes
	log.Infof("Deploying  code for id: %s", a.Id)
	if err := a.Restart(); err != nil {
		log.Errorf("Error deploying image: %v", err)
		events.Send(a.Id, events.Event{Body: "An error occurred during the push process: " + err.Error(), Type: events.Message})
		return
	}
	events.Send(a.Id, events.Event{Body: "Deploy complete", Type: events.Message})
}

func Build(w http.ResponseWriter, r *http.Request) {
	// get or create app
	a, err := getApp(w, r)
	if err != nil {
		return
	}

	deploy, err := strconv.ParseBool(r.FormValue("deploy"))
	if err != nil {
		deploy = false
	}

	go build(a, deploy)
}
