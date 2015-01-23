package app

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/myodc/playground-server/server/docker"
	"github.com/myodc/playground-server/server/events"
	"github.com/myodc/playground-server/server/kubernetes"
	"github.com/myodc/playground-server/server/store"
	log "github.com/cihub/seelog"
)

var (
	baseImage       = `FROM myodc/playground-base`
	namespace       = "playground:apps"
	statusNamespace = "playground:apps:status"
	nameRe          = regexp.MustCompilePOSIX("^[a-z][a-z0-9-]+")
)

func Create(app *App) error {
	if !nameRe.MatchString(app.Id) {
		return fmt.Errorf("App Id invalid. Must match %s", nameRe.String())
	}

	exists, err := store.Exists(namespace, app.Id)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("App already exists")
	}

	if err := Update(app); err != nil {
		return err
	}

	return app.UpdateStatus(&Info{AppId: app.Id, Status: "Created", Timestamp: time.Now()})
}

func Update(app *App) error {
	if !nameRe.MatchString(app.Id) {
		return fmt.Errorf("App Id invalid. Must match %s", nameRe.String())
	}

	if app.Source == nil {
		return fmt.Errorf("App source not set")
	}

	if app.Created.IsZero() {
		app.Created = time.Now()
	}

	if app.Config == nil {
		app.Config = &Config{
			ContainerPort: 8080,
			NumInstances:  1,
		}
	}

	if len(app.Source.Image) > 0 {
		app.Image = app.Source.Image
	}

	app.Updated = time.Now()

	b, err := json.Marshal(app)
	if err != nil {
		return err
	}

	return store.Put(namespace, app.Id, b)
}

func Delete(id string) error {
	// Remove running app
	kubernetes.Delete(id)
	return store.Del(namespace, id)
}

func Read(id string) (*App, error) {
	b, err := store.Get(namespace, id)
	if err != nil {
		return nil, err
	}
	var app *App
	err = json.Unmarshal(b, &app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func List(offset, limit int) ([]*App, error) {
	results, err := store.Range(namespace, offset, limit)
	if err != nil {
		return nil, err
	}
	var apps []*App
	for _, result := range results {
		var app *App
		err := json.Unmarshal(result, &app)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func Logs(id, container string, out io.Writer) error {
	return kubernetes.Logs(id, container, out)
}

func Status(id string) (*Info, error) {
	if !nameRe.MatchString(id) {
		return nil, fmt.Errorf("App Id invalid. Must match %s", nameRe.String())
	}

	exists, err := store.Exists(namespace, id)
	if err != nil {
		log.Errorf("Error finding app %s: %v", id, err)
		return nil, fmt.Errorf("Error retrieving status")
	}

	if !exists {
		return nil, fmt.Errorf("Does not exist")
	}

	b, err := store.Get(statusNamespace, id)
	if err != nil {
		log.Errorf("Error getting status for %s: %v", id, err)
		return nil, fmt.Errorf("Error retrieving status")
	}

	var status *Info
	err = json.Unmarshal(b, &status)
	if err != nil {
		log.Errorf("Error getting status for %s: %v", id, err)
		return nil, fmt.Errorf("Error retrieving status")
	}

	return status, nil
}

// Build a piece of code as an image
func (a *App) Build() error {
	// get app from store
	// Build {Code, GitRepo, Dockerfile}

	// TODO: set app status to building

	if a.Source == nil {
		return fmt.Errorf("App source does not exist")
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Building",
		Reason:  "Build executed",
		Message: "Building image for app",
	})

	var err error
	switch {
	case a.Source.Code != nil:
		err = buildCode(a)
	case a.Source.GitRepo != nil:
		err = buildGitRepo(a)
	case len(a.Source.Dockerfile) > 0:
		err = buildDockerFile(a)
	case len(a.Source.Image) > 0:
		a.Image = a.Source.Image
		// update status
		a.UpdateStatus(&Info{
			Status:  "Built",
			Reason:  "Build executed",
			Message: "Build not needed, image specified",
		})
		return nil
	default:
		// update status
		a.UpdateStatus(&Info{
			Status:  "Failed",
			Reason:  "Invalid source",
			Message: "Invalid Source specified for app",
		})
		return fmt.Errorf("Cannot build, correct source not specified")
	}

	if err != nil {
		return err
	}

	a.Image = fmt.Sprintf("%s:%s", docker.Image(a.Id), "latest")

	if err := Update(a); err != nil {
		// update status
		a.UpdateStatus(&Info{
			Status:  "Failed",
			Reason:  err.Error(),
			Message: "Failed to update app state",
		})
		return err
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Built",
		Reason:  "Build executed",
		Message: "Built image for app",
	})

	return nil
}

// Push sends the image to the docker registry
func (a *App) Push() error {
	if len(a.Image) == 0 {
		return fmt.Errorf("App image not set")
	}

	if a.Source == nil {
		return fmt.Errorf("App source does not exist")
	}

	// Don't push when the source is an image
	if len(a.Source.Image) > 0 {
		return nil
	}

	in, out := io.Pipe()

	// make the output available for streaming
	go events.Receive(a.Id, in)

	// update status
	a.UpdateStatus(&Info{
		Status:  "Pushing",
		Reason:  "Push executed",
		Message: "Pushing to registry",
	})

	if err := docker.Push(a.Id, "latest", true, out); err != nil {
		// update status
		a.UpdateStatus(&Info{
			Status:  "Failed",
			Reason:  err.Error(),
			Message: "Failed pushing to registry",
		})
		return err
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Pushed",
		Reason:  "Push executed",
		Message: "Pushed to registry",
	})

	return nil
}

func (a *App) Restart() error {
	a.Stop()
	return a.Start()
}

// Start pushes the build into kubernetes
func (a *App) Start() error {
	if len(a.Image) == 0 {
		return fmt.Errorf("App image not set")
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Starting",
		Reason:  "Start executed",
		Message: "Starting app",
	})

	// start service
	service, err := kubernetes.Create(a.Id, &kubernetes.ContainerConfig{
		ContainerPort: a.Config.ContainerPort,
		Image:         a.Image,
		NumInstances:  a.Config.NumInstances,
		Labels: map[string]string{
			"name":  a.Id,
			"type":  "playground",
			"proxy": "true",
		},
	})
	if err != nil {
		a.UpdateStatus(&Info{
			Status:  "Failed",
			Reason:  err.Error(),
			Message: "Failed to start app",
		})
		return err
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Started",
		Reason:  "Start executed",
		Message: fmt.Sprintf("App has been started: %v", *service),
	})

	return nil
}

// Remove deletes a build running on kubernetes
func (a *App) Stop() error {
	// update status
	a.UpdateStatus(&Info{
		Status:  "Stopping",
		Reason:  "Stop executed",
		Message: "Stopping app",
	})

	// start service
	if err := kubernetes.Delete(a.Id); err != nil {
		a.UpdateStatus(&Info{
			Status:  "Failed",
			Reason:  err.Error(),
			Message: "Failed to stop app",
		})
		return err
	}

	// update status
	a.UpdateStatus(&Info{
		Status:  "Stopped",
		Reason:  "Stop executed",
		Message: "App has been stopped",
	})

	return nil
}

// Update Status
func (a *App) UpdateStatus(info *Info) error {
	info.AppId = a.Id
	info.Timestamp = time.Now()

	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	events.Send(a.Id, events.Event{Body: info.Status, Type: events.Status})
	return store.Put(statusNamespace, a.Id, b)
}
