package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/myodc/playground-server/server/app"
)

func getApp(w http.ResponseWriter, r *http.Request) (*app.App, error) {
	id := r.FormValue("id")
	tsk := r.FormValue("app")

	if len(id) == 0 && len(tsk) == 0 {
		msg := "App or ID not found"
		http.Error(w, msg, http.StatusBadRequest)
		return nil, errors.New(msg)
	}

	if len(id) > 0 {
		// read existing app
		aapp, err := app.Read(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		return aapp, nil
	}

	// create app if it does not exist
	var aapp *app.App
	if err := json.Unmarshal([]byte(tsk), &aapp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	if err := app.Create(aapp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return aapp, nil
}
