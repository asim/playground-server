package app

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/myodc/playground-server/server/docker"
	"github.com/myodc/playground-server/server/events"
)

func buildDockerFile(app *App) error {
	dir, err := ioutil.TempDir("", "playground")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	// Write Dockefile
	dockerFile := filepath.Join(dir, "Dockerfile")

	f, err := os.Create(dockerFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(app.Source.Dockerfile); err != nil {
		return err
	}

	in, out := io.Pipe()

	// make the output available for streaming
	go events.Receive(app.Id, in)

	// blocking
	return docker.Build(app.Id, "latest", dir, out)
}
