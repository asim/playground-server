package handler

import (
	"encoding/json"
	"net/http"

	"github.com/myodc/playground-server/server/app"
)

// Status returns app status.
/*
	"id": "foo"
*/
func Status(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		http.Error(w, "Require app Id", http.StatusBadRequest)
		return
	}

	status, err := app.Status(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
