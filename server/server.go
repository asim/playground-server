package server

import (
	"log"
	"net/http"

	"github.com/myodc/playground-server/server/handler"
)

type server struct{}

func init() {
	// Code: Short lived
	http.HandleFunc("/code/share", handler.Share)
	http.HandleFunc("/code/load", handler.Load)
	http.HandleFunc("/code/run", handler.Run)

	// Tasks: Long Lived
	http.HandleFunc("/apps/list", handler.List)
	http.HandleFunc("/apps/create", handler.Create)
	http.HandleFunc("/apps/delete", handler.Delete)
	http.HandleFunc("/apps/update", handler.Update)
	http.HandleFunc("/apps/read", handler.Read)

	// Deployment
	http.HandleFunc("/apps/logs", handler.Logs)
	http.HandleFunc("/apps/build", handler.Build)
	http.HandleFunc("/apps/status", handler.Status)
	http.HandleFunc("/apps/start", handler.Start)
	http.HandleFunc("/apps/stop", handler.Stop)

	// Event stream
	http.HandleFunc("/events", handler.Events)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		return
	}

	http.DefaultServeMux.ServeHTTP(w, r)
}

func Run(address string) {
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, &server{}); err != nil {
		panic(err.Error())
	}
}
