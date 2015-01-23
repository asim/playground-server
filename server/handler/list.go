package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/myodc/playground-server/server/app"
)

// List returns a list of long running apps
func List(w http.ResponseWriter, r *http.Request) {
	of := r.FormValue("offset")
	li := r.FormValue("limit")

	offset, err := strconv.Atoi(of)
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(li)
	if err != nil {
		limit = 20
	}

	apps, err := app.List(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var b []byte
	if len(apps) == 0 {
		b = []byte(`{}`)
	} else {
		var err error
		b, err = json.Marshal(map[string][]*app.App{"apps": apps})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	return
}
