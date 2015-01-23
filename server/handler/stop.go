package handler

import (
	"net/http"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
)

// Stop will remove a long running app from kubernetes
/*
	"id": foo
*/
func stop(a *app.App) {
	if err := a.Stop(); err != nil {
		events.Send(a.Id, events.Event{Body: err.Error(), Type: events.Error})
		return
	}

	events.Send(a.Id, events.Event{Body: "Remove complete", Type: events.Message})
}

func Stop(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		events.Send(id, events.Event{Body: "Id cannot be blank", Type: events.Error})
		http.Error(w, "Id cannot be blank", http.StatusBadRequest)
		return
	}

	a, err := app.Read(id)
	if err != nil {
		events.Send(id, events.Event{Body: "App not found", Type: events.Error})
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}

	go stop(a)
}
