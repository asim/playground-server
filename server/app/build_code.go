package app

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/myodc/playground-server/server/docker"
	"github.com/myodc/playground-server/server/events"
	"github.com/myodc/playground-server/server/lang"
)

func buildCode(app *App) error {
	// TODO: create a build status updater of some kind
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

	if _, err := f.WriteString(baseImage); err != nil {
		return err
	}

	// Get code file extension
	ext, err := lang.ToExt(app.Source.Code.Lang)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("main.%s", ext)
	filePath := path.Join(dir, fileName)

	// Write code to file
	f2, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f2.Close()

	if _, err := f2.WriteString(app.Source.Code.Text); err != nil {
		return err
	}

	in, out := io.Pipe()

	// make the output available for streaming
	go events.Receive(app.Id, in)

	// blocking
	return docker.Build(app.Id, "latest", dir, out)
}
