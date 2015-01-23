package handler

import (
	"fmt"
	"net/http"

	"github.com/myodc/playground-server/server/events"
	log "github.com/cihub/seelog"
	"github.com/gorilla/websocket"
)

func Events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")

	if len(id) == 0 {
		id = "*"
	}

	ch := events.Subscribe(id)

	log.Infof("New Event subscriber (%s)", id)

	defer func() {
		log.Infof("Cleaning up connection %s", id)
		events.Unsubscribe(id, ch)
		conn.Close()
	}()

	for {
		select {
		case e := <-ch:
			err := conn.WriteJSON(e)
			if err != nil {
				log.Error(w, fmt.Sprintf("error sending ws message: %v", err.Error()))
				return
			}
		}
	}

	log.Infof("Event Streaming Complete for %s", id)
}
