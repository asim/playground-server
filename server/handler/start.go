package handler

import (
	"net/http"
	"strconv"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
	log "github.com/cihub/seelog"
)

// Start will create a long running app
/*

	"id": foo

		OR

	"app": {
                "id": "foo",
                "description": "Some thing",
                "source": {
                        "dode": {
                                "lang": "ruby",
                                "text": "puts 'hello world'",
			}
		}
        }
	"build": true

*/
func start(a *app.App, build bool) {
	// build app
	if build {
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
	}

	if err := a.Start(); err != nil {
		events.Send(a.Id, events.Event{Body: err.Error(), Type: events.Error})
		return
	}

	events.Send(a.Id, events.Event{Body: "Deploy complete", Type: events.Message})
}

func Start(w http.ResponseWriter, r *http.Request) {
	// get or create app
	a, err := getApp(w, r)
	if err != nil {
		return
	}

	build, err := strconv.ParseBool(r.FormValue("build"))
	if err != nil {
		build = false
	}

	go start(a, build)
}
