package docker

import (
	"os"
	"path/filepath"
)

var (
	repository = "playground"
)

func registry() string {
	host := os.Getenv("PLAYGROUND_REGISTRY_SERVICE_HOST")
	port := os.Getenv("PLAYGROUND_REGISTRY_SERVICE_PORT")

	if len(host) > 0 && len(port) > 0 {
		return filepath.Join(host+":"+port, repository)
	}

	return filepath.Join("localhost:5000", repository)
}
