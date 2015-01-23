package handler

import (
	"net/http"
	"time"

	"github.com/myodc/playground-server/server/code"
	"github.com/myodc/playground-server/server/events"
	log "github.com/cihub/seelog"
)

// Run starts a short lived app
/*
{
	"id": "foo"
}
*/
func Run(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	log.Infof("Running code for id: %s", id)
	status, err := code.Run(id, &code.Code{
		Lang: r.FormValue("lang"),
		Text: r.FormValue("text"),
	}, 10*time.Second)

	if err != nil {
		log.Errorf("Error running code: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events.Send(id, events.Event{Body: status, Type: events.Message})
}
