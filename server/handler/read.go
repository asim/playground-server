package handler

import (
	"encoding/json"
	"net/http"

	"github.com/myodc/playground-server/server/app"
)

// Read returns a app.
/*
	"id": "foo"
*/
func Read(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		http.Error(w, "Require app Id", http.StatusBadRequest)
		return
	}

	aapp, err := app.Read(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(aapp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
