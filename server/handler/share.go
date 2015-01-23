package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/myodc/playground-server/server/code"
	log "github.com/cihub/seelog"
)

// Share saves code and returns a unique id for code
/*
{
	"lang": "golang",
	"text": "package fmt..."
}
*/
func Share(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	text := r.FormValue("text")

	id := code.GenShortId()
	err := code.Save(id, &code.Code{Lang: lang, Text: text})
	if err != nil {
		log.Errorf("Error saving code %s: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(map[string]string{
		"id": id,
	})
	fmt.Fprint(w, string(b))
}
