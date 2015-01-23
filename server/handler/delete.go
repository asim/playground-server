package handler

import (
	"net/http"

	"github.com/myodc/playground-server/server/app"
	"github.com/myodc/playground-server/server/events"
)

// Delete removes a app.
/*
	"id": "foo"
*/
func Delete(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		http.Error(w, "Require app Id", http.StatusBadRequest)
		return
	}

	go func() {
		err := app.Delete(id)
		if err != nil {
			events.Send(id, events.Event{Body: "Error deleting app: " + err.Error(), Type: events.Message})
			return
		}

		events.Send(id, events.Event{Body: "App deleted", Type: events.Message})
	}()
}
