package handler

import (
	"fmt"
	"net/http"

	"github.com/myodc/playground-server/server/code"
)

// Load saved code
/*
{
	"id": "foo"
}
*/
func Load(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if len(id) == 0 {
		http.Error(w, "Code not found", 404)
		return
	}

	raw, _ := code.Load(id)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, raw)
}
