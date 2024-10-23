package runtime

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"io"
)

type DockerRuntime struct {
	name    string
	image   string
	version string
}

func NewDockerRuntime(name, image, version string) *DockerRuntime {
	return &DockerRuntime{
		name:    name,
		image:   image,
		version: version,
	}
}

func (dr *DockerRuntime) Init(fn *function.Function) error {
	var err error

	ctx := context.Background()

	// Check if image exists, pull if not
	_, _, err = dockerClient.ImageInspectWithRaw(ctx, dr.image)
	if client.IsErrNotFound(err) {
		fmt.Println("Pulling image...")
		reader, err := dockerClient.ImagePull(ctx, dr.image, image.PullOptions{})
		if err != nil {
			return err
		}
		_, _ = io.ReadAll(reader)
		_ = reader.Close()
	}
	if fn != nil {
		containerId, err := dr.createContainer(ctx, fn)
		if err != nil {
			return err
		}
		fn.ContainerID = containerId
	}
	return nil
}

func (dr *DockerRuntime) createContainer(ctx context.Context, fn *function.Function) (string, error) {
	cm := NewContainerManager(fn, dr.GetImage())
	return cm.getOrCreateContainer(ctx)
}

func (dr *DockerRuntime) Shutdown(fn *function.Function) error {
	// Stop and remove container, gracefully
	cm := NewContainerManager(fn, dr.GetImage())
	return cm.Cleanup(true)
}

func (dr *DockerRuntime) GetImage() string {
	return dr.image
}

func (dr *DockerRuntime) GetCmd(event event.Event, fn function.Function) []string {
	cmd := []string{fn.Handler}
	payload := event.GetPayload()
	if payload != "" {
		cmd = append(cmd, payload)
	}
	return cmd
}

func Execute(ctx context.Context, r Runtime, event *event.Event, fn *function.Function) (string, error) {
	cmd := r.GetCmd(event, fn)
	cm := NewContainerManager(fn, r.GetImage())
	stdout, stderr, err := cm.RunCommand(ctx, cmd)
	fn.ContainerID = cm.containerID
	fn.LastRun = cm.lastUsed

	if err != nil {
		return "", err
	}
	if stderr != "" {
		return "", fmt.Errorf(stderr)
	}
	return stdout, nil
}
