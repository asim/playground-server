package code

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/myodc/playground-server/server/docker"
	"github.com/myodc/playground-server/server/events"
	"github.com/myodc/playground-server/server/store"
)

type Code struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}

var (
	alphanum        = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	templateProject = "https://github.com/myodc/playground-server/server"
	namespace       = "playground:code"
)

func GenShortId() string {
	bytes := make([]byte, 10)
	for i := 0; i < 1000; i++ {
		rand.Read(bytes)
		for i, b := range bytes {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		}
		id := string(bytes)
		exists, err := store.Exists(namespace, id)
		if err != nil {
			continue
		}
		if !exists {
			return id
		}
	}

	return "123zYxWvuTsR"
}

// Load a piece of code
func Load(id string) (string, error) {
	code, err := store.Get(namespace, id)
	if err != nil {
		return "", err
	}
	return string(code), nil
}

// Save a piece of code for sharing
func Save(id string, code *Code) error {
	b, err := json.Marshal(code)
	if err != nil {
		return err
	}
	return store.Put(namespace, id, b)
}

// Run a one off short lived app
func Run(id string, code *Code, duration time.Duration) (string, error) {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()
	defer outReader.Close()
	defer errReader.Close()

	c := docker.CodeContainer(code.Lang, code.Text)

	c.StdOut = outWriter
	c.StdErr = errWriter

	go events.Receive(id, outReader)
	go events.Receive(id, errReader)

	status, err := c.Run(duration)
	if err != nil {
		return "", err
	}

	switch status {
	case -1:
		return "[Program took too long]", nil
	default:
		return fmt.Sprintf("[Program exited: status %d]", status), nil
	}
}
