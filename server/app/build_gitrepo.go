package app

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/myodc/playground-server/server/docker"
	"github.com/myodc/playground-server/server/events"
)

func buildGitRepo(app *App) error {
	// TODO: create a build status updater of some kind

	dir, err := ioutil.TempDir("", "playground")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	repo := filepath.Join(dir, "repo")

	// check url
	if len(app.Source.GitRepo.Url) == 0 {
		return fmt.Errorf("Git URL cannot be blank")
	}

	branch := app.Source.GitRepo.Branch
	if len(branch) == 0 {
		branch = "master"
	}

	in, out := io.Pipe()

	cmd := exec.Command("git", "clone", "-b", branch, app.Source.GitRepo.Url, repo)
	cmd.Stdout = out
	cmd.Stderr = out

	// make the output available for streaming
	go events.Receive(app.Id, in)

	if err := cmd.Run(); err != nil {
		return err
	}

	// blocking
	return docker.Build(app.Id, "latest", repo, out)
}
