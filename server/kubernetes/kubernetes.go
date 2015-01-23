package kubernetes

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	log "github.com/cihub/seelog"
)

type ContainerConfig struct {
	Image         string
	ContainerPort int
	NumInstances  int
	Labels        map[string]string
}

type Service struct {
	Name   string
	IP     string
	Port   int
	Status string
}

var (
	defaultPort      = 8080
	defaultNamespace = "default"
)

func newClient() (*client.Client, error) {
	config := &client.Config{
		Host:     "http://127.0.0.1:8080",
		Insecure: true,
		Username: os.Getenv("PLAYGROUND_KUBE_USER"),
		Password: os.Getenv("PLAYGROUND_KUBE_PASS"),
	}

	if host := os.Getenv("PLAYGROUND_KUBE_HOST"); len(host) > 0 {
		config.Host = host
	}

	return client.New(config)
}

func EventWatcher() {
	client, err := newClient()
	if err != nil {
		log.Error(err)
		return
	}

	el, err := client.Events(defaultNamespace).List(labels.OneTermEqualSelector("involvedObject.Name", "type"), labels.OneTermEqualSelector("involvedObject.Name", "playground"))
	if err != nil {
		log.Error(err)
		return
	}
	rv := el.ListMeta.ResourceVersion
	log.Debug(el)
	w, err := client.Events(defaultNamespace).Watch(labels.OneTermEqualSelector("involvedObject.Name", "type"), labels.OneTermEqualSelector("involvedObject.Name", "playground"), rv)
	if err != nil {
		log.Error(err)
		return
	}
	for {
		select {
		case e := <-w.ResultChan():
			log.Debugf("Result %v", e)
		}
	}
}

func replCtrlFromConfig(name string, config *ContainerConfig) *api.ReplicationController {
	if config == nil {
		config = &ContainerConfig{}
	}

	if config.ContainerPort == 0 {
		config.ContainerPort = defaultPort
	}

	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}

	config.Labels["name"] = name

	container := api.Container{
		Name:  name,
		Image: config.Image,
		Ports: []api.Port{
			api.Port{
				ContainerPort: config.ContainerPort,
			},
		},
		ImagePullPolicy: api.PullAlways,
	}

	return &api.ReplicationController{
		api.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1beta1",
		},
		api.ObjectMeta{
			Name:   name,
			Labels: config.Labels,
		},
		api.ReplicationControllerSpec{
			Replicas: config.NumInstances,
			Selector: map[string]string{
				"name": name,
			},
			Template: &api.PodTemplateSpec{
				api.ObjectMeta{
					Name:   name,
					Labels: config.Labels,
				},
				api.PodSpec{
					Containers: []api.Container{
						container,
					},
				},
			},
		},
		api.ReplicationControllerStatus{},
	}
}

func Create(name string, config *ContainerConfig) (*Service, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	repl := replCtrlFromConfig(name, config)
	repl, err = client.ReplicationControllers(defaultNamespace).Create(repl)
	if err != nil {
		return nil, err
	}

	// create service config
	service := &api.Service{
		api.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1beta1",
		},
		api.ObjectMeta{
			Name:   name,
			Labels: config.Labels,
		},
		api.ServiceSpec{
			Port: defaultPort,
			Selector: map[string]string{
				"name": name,
			},
			ContainerPort: util.NewIntOrStringFromInt(config.ContainerPort),
		},
		api.ServiceStatus{},
	}

	service, err = client.Services(defaultNamespace).Create(service)
	if err != nil {
		return nil, err
	}

	// save pod and service state
	return &Service{
		Name:   name,
		IP:     service.Spec.PortalIP,
		Port:   service.Spec.Port,
		Status: "Pending",
	}, nil
}

func Update(name string, config *ContainerConfig, out io.Writer) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	oldRc, err := client.ReplicationControllers(defaultNamespace).Get(name)
	if err != nil {
		return err
	}
	newRc := replCtrlFromConfig(name, config)

	updater := kubectl.NewRollingUpdater(defaultNamespace, client)

	var hasLabel bool
	for key, oldValue := range oldRc.Spec.Selector {
		if newValue, ok := newRc.Spec.Selector[key]; ok && newValue != oldValue {
			hasLabel = true
			break
		}
	}
	if !hasLabel {
		return errors.New("Selector does not match existing replicas")
	}
	// TODO: handle resizes during rolling update
	if newRc.Spec.Replicas == 0 {
		newRc.Spec.Replicas = oldRc.Spec.Replicas
	}
	return updater.Update(out, oldRc, newRc, time.Minute, time.Second*3, time.Minute*5)
}

func Delete(name string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	var errs []string

	if err := client.Services(defaultNamespace).Delete(name); err != nil {
		errs = append(errs, "service error: "+err.Error())
	}

	oldRc, err := client.ReplicationControllers(defaultNamespace).Get(name)
	if err != nil {
		errs = append(errs, "replication controller error: "+err.Error())
	} else {
		oldRc.Spec.Replicas = 0
		if _, err := client.ReplicationControllers(defaultNamespace).Update(oldRc); err != nil {
			errs = append(errs, "replication controller error: "+err.Error())
		}

		time.Sleep(time.Second * 10)
		if err := client.ReplicationControllers(defaultNamespace).Delete(name); err != nil {
			errs = append(errs, "replication controller error: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func Logs(id, container string, out io.Writer) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	pod, err := client.Pods(defaultNamespace).Get(id)
	if err != nil {
		return err
	}

	if len(container) == 0 {
		if len(pod.Spec.Containers) != 1 {
			return errors.New("<container> is required for pods with multiple containers")
		}

		// Get logs for the only container in the pod
		container = pod.Spec.Containers[0].Name
	}

	readCloser, err := client.RESTClient.Get().
		Prefix("proxy").
		Resource("minions").
		Name(pod.Status.Host).
		Suffix("containerLogs", defaultNamespace, id, container).
		Param("follow", strconv.FormatBool(false)).
		Stream()

	if err != nil {
		return err
	}

	defer readCloser.Close()

	_, err = io.Copy(out, readCloser)
	if err != nil {
		return err
	}

	return nil
}
