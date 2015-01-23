package handler

import (
	"net/http"

	"github.com/myodc/playground-server/server/app"
)

// Logs returns app logs.
/*
	"id": "foo"
	"container": "foo" [optional]
*/

func Logs(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		http.Error(w, "Require app Id", http.StatusBadRequest)
		return
	}

	container := r.FormValue("container")

	err := app.Logs(id, container, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
