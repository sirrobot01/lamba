package runtime

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
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

func (dr *DockerRuntime) Init() error {
	var err error

	ctx := context.Background()

	// Check if image exists, pull if not
	_, _, err = dockerClient.ImageInspectWithRaw(ctx, dr.image)
	if client.IsErrNotFound(err) {
		fmt.Println("Pulling image...")
		reader, err := dockerClient.ImagePull(ctx, dr.image, image.PullOptions{})
		if err != nil {
			panic(err)
		}
		io.ReadAll(reader)
		reader.Close()
	}
	return nil
}

func (dr *DockerRuntime) Shutdown() error {
	return nil
}

func (dr *DockerRuntime) GetCmd(event event.InvokeEvent, fn function.Function) []string {
	cmd := []string{fn.Handler}
	payload := string(event.Payload)
	if payload != "" {
		cmd = append(cmd, payload)
	}
	return cmd
}

func (dr *DockerRuntime) Execute(ctx context.Context, event event.InvokeEvent, fn function.Function) ([]byte, error) {
	cmd := dr.GetCmd(event, fn)
	containerConfig := &container.Config{
		Image: dr.image,
		Cmd:   cmd,
		Tty:   false,
	}
	resp, err := dockerClient.ContainerCreate(ctx, containerConfig, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app", fn.Name),
		},
	}, nil, nil, fn.Name)
	if err != nil {
		return nil, err
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}
	statusCh, errCh := dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-statusCh:
	}

	out, err := dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: false})
	if err != nil {
		return nil, err
	}

	// Read the output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, out)
	if err != nil {
		return nil, err
	}

	output := buf.String()

	// Remove the container
	if err := dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		return nil, err
	}
	return []byte(output), nil

}
