package docker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/utils"
	dcli "github.com/fsouza/go-dockerclient"
)

func newClient() (*dcli.Client, error) {
	endpoint := os.Getenv("DOCKER_HOST")
	tls := os.Getenv("DOCKER_TLS_VERIFY")
	certPath := os.Getenv("DOCKER_CERT_PATH")

	if len(endpoint) == 0 {
		endpoint = "unix:///var/run/docker.sock"
	}

	if len(tls) > 0 && len(certPath) > 0 {
		return dcli.NewTLSClient(
			endpoint,
			filepath.Join(certPath, "cert.pem"),
			filepath.Join(certPath, "key.pem"),
			filepath.Join(certPath, "ca.pem"),
		)
	}

	return dcli.NewClient(endpoint)
}

func Build(name, tag, dir string, out io.Writer) error {
	options := &archive.TarOptions{
		Compression: archive.Uncompressed,
	}

	context, err := archive.TarWithOptions(dir, options)
	if err != nil {
		return err
	}

	var in io.Reader
	if context != nil {
		sf := utils.NewStreamFormatter(false)
		in = utils.ProgressReader(context, 0, out, sf, true, "", "Sending build context to Docker daemon")
	}

	opts := dcli.BuildImageOptions{
		Name:           fmt.Sprintf("%s/%s:%s", registry(), name, tag),
		InputStream:    in,
		OutputStream:   out,
		RmTmpContainer: true,
	}

	// new docker client
	client, err := newClient()
	if err != nil {
		return err
	}

	// build docker image
	if err := client.BuildImage(opts); err != nil {
		return err
	}

	return nil
}

func Image(name string) string {
	return fmt.Sprintf("%s/%s", registry(), name)
}

func Exists(image, tag string) bool {
	// new docker client
	client, err := newClient()
	if err != nil {
		return false
	}

	_, err = client.InspectImage(image + ":" + tag)
	if err != nil {
		return false
	}
	return true
}

func Pull(image, tag string, out io.Writer) error {
	// new docker client
	client, err := newClient()
	if err != nil {
		return err
	}

	return client.PullImage(dcli.PullImageOptions{
		Repository:   image,
		Tag:          tag,
		OutputStream: out,
	}, dcli.AuthConfiguration{})
}

func Push(name, tag string, rmImage bool, out io.Writer) error {
	// new docker client
	client, err := newClient()
	if err != nil {
		return err
	}

	popts := dcli.PushImageOptions{
		Name:         fmt.Sprintf("%s/%s", registry(), name),
		Tag:          tag,
		Registry:     registry(),
		OutputStream: out,
	}

	if err := client.PushImage(popts, dcli.AuthConfiguration{}); err != nil {
		return err
	}

	if rmImage {
		return client.RemoveImage(fmt.Sprintf("%s/%s:%s", registry(), name, tag))
	}

	return nil
}
