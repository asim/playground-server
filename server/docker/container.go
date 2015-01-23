package docker

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/myodc/playground-server/server/lang"
	log "github.com/cihub/seelog"
	dcli "github.com/fsouza/go-dockerclient"
)

type Code struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}

type Container struct {
	Id     string
	Dir    string
	Code   *Code
	StdOut io.Writer
	StdErr io.Writer
	Uid    int
}

var (
	alphanum  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	baseImage = "myodc/playground-base"
	proc      *dcli.Client
	pool      *uidPool
)

func Init() {
	initProc()
	initPool()
}

func initProc() {
	if proc != nil {
		return
	}

	client, err := newClient()
	if err != nil {
		panic(err.Error())
	}

	proc = client
}

func initPool() {
	if pool != nil {
		return
	}

	pool = newUidPool(20000, 25000)
}

func CodeContainer(lang, text string) *Container {
	if proc == nil {
		initProc()
	}

	return &Container{
		Code: &Code{
			Lang: lang,
			Text: text,
		},
		Uid: 10000,
	}
}

func (c *Container) Run(timeout time.Duration) (int, error) {
	log.Info("Creating code directory")
	dir, err := ioutil.TempDir("", "playground-")
	if err != nil {
		return 0, err
	}
	c.Dir = dir

	log.Info("Creating source file")
	srcFile, err := c.createSrcFile()
	if err != nil {
		return 0, err
	}

	if pool == nil {
		Init()
	}

	log.Info("Reserving uid")
	uid, err := pool.Get()
	if err != nil {
		return 0, err
	}
	log.Infof("Got uid %d", uid)
	c.Uid = uid

	// check if image exists, otherwise pull
	if !Exists(baseImage, "latest") {
		err := Pull(baseImage, "latest", c.StdOut)
		if err != nil {
			return 0, err
		}
	}

	log.Info("Creating container")
	if err := c.createContainer(srcFile); err != nil {
		return 0, err
	}

	log.Info("Starting container")
	if err := c.startContainer(); err != nil {
		return 0, err
	}
	defer c.cleanup()

	log.Info("Streaming container logs")
	go func() {
		if err := c.streamLogs(); err != nil {
			log.Errorf("%v", err)
		}
	}()

	log.Info("Waiting for container to finish")
	killed, status := c.wait(timeout)
	if killed {
		log.Errorf("Container exited with status %d", status)
		return -1, nil
	}

	return 0, nil
}

func (c *Container) createSrcFile() (string, error) {
	ext, err := lang.ToExt(c.Code.Lang)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("program.%s", ext)
	filePath := path.Join(c.Dir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(c.Code.Text); err != nil {
		return "", err
	}

	return fileName, nil
}

func (c *Container) createContainer(srcFile string) error {
	uidStr := strconv.Itoa(c.Uid)
	opts := dcli.CreateContainerOptions{
		Config: &dcli.Config{
			CPUShares: 1,
			Memory:    50e6,
			Tty:       true,
			OpenStdin: false,
			Volumes: map[string]struct{}{
				"/code": {},
			},
			Cmd:   []string{path.Join("/code", srcFile), uidStr},
			Image: baseImage,
		},
	}

	container, err := proc.CreateContainer(opts)
	if err != nil {
		return err
	}

	c.Id = container.ID
	return nil
}

func (c *Container) startContainer() error {
	if c.Id == "" {
		return errors.New("Can't start a container before it is created")
	}

	hostConfig := &dcli.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/code", c.Dir)},
	}
	if err := proc.StartContainer(c.Id, hostConfig); err != nil {
		return err
	}

	return nil
}

func (c *Container) wait(timeout time.Duration) (bool, int) {
	statusCh := make(chan int, 1)

	go func() {
		if status, err := proc.WaitContainer(c.Id); err != nil {
			log.Infof("%v", err)
		} else {
			statusCh <- status
		}
	}()

	var killed bool

	for {
		select {
		case status := <-statusCh:
			log.Infof("Container %s exited by itself", c.Id)
			return killed, status
		case <-time.After(timeout):
			log.Infof("Container %s timed out, killing", c.Id)
			if err := proc.StopContainer(c.Id, 0); err != nil {
				log.Infof("%v", err)
			}
			killed = true
		}
	}
}

func (c *Container) cleanup() {
	log.Infof("Removing container %s", c.Id)
	if err := proc.RemoveContainer(dcli.RemoveContainerOptions{ID: c.Id}); err != nil {
		log.Errorf("Couldn't remove container %s (%v)", c.Id, err)
	}

	log.Infof("Removing code dir %s", c.Dir)
	if err := os.RemoveAll(c.Dir); err != nil {
		log.Errorf("Couldn't remove temp dir %s (%v)", c.Dir, err)
	}

	if pool == nil {
		Init()
	}

	log.Infof("Releasing uid %d", c.Uid)
	if err := pool.Put(c.Uid); err != nil {
		log.Errorf("Couldn't release uid %d (%v)", c.Uid, err)
	}
	c.Uid = 0
}

func (c *Container) streamLogs() error {
	if len(c.Id) == 0 {
		return errors.New("Can't attach to a container before it is created")
	}

	opts := dcli.LogsOptions{
		Container:    c.Id,
		OutputStream: c.StdOut,
		ErrorStream:  c.StdErr,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
	}
	if err := proc.Logs(opts); err != nil {
		return err
	}

	return nil
}
