package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"
	"io"
	"sync"
	"time"
)

type Engine struct {
	image         string
	containerID   string
	containerName string
	lastUsed      time.Time
	mutex         sync.Mutex
	CodePath      string
	client        *client.Client
}

func NewEngine(containerId, name, image, codePath string) *Engine {
	cl, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &Engine{
		client:        cl,
		image:         image,
		containerID:   containerId,
		containerName: name,
		CodePath:      codePath,
	}
}

func (e *Engine) UpdateEngine(containerId, name, codePath string) {
	e.containerID = containerId
	e.containerName = name
	e.CodePath = codePath
}

func (e *Engine) PullImage(ctx context.Context) error {
	_, _, err := e.client.ImageInspectWithRaw(ctx, e.image)
	if client.IsErrNotFound(err) {
		log.Debug().Msgf("Pulling image: %s", e.image)
		reader, err := e.client.ImagePull(ctx, e.image, image.PullOptions{})
		if err != nil {
			return err
		}
		_, _ = io.ReadAll(reader)
		_ = reader.Close()
	}
	return nil
}

func (e *Engine) GetOrCreateContainer(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if container exists and is recent
	if e.containerID != "" {
		// Check if container is still running
		_, err := e.client.ContainerInspect(ctx, e.containerID)
		if err == nil {
			e.lastUsed = time.Now()
			return e.containerID, nil
		}
	}

	// Check if image exists, pull if not
	err := e.PullImage(ctx)
	if err != nil {
		return "", err
	}

	// Create new container
	hostConf := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app", e.CodePath),
		},
	}
	resp, err := e.client.ContainerCreate(ctx,
		&container.Config{
			Image:      e.image,
			Cmd:        []string{"tail", "-f", "/dev/null"}, // Keep container running
			Tty:        true,
			WorkingDir: "/app",
		},
		hostConf, nil, nil, e.containerName)
	if err != nil {
		return "", err
	}
	// Start container
	if err := e.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	e.containerID = resp.ID
	e.lastUsed = time.Now()
	return resp.ID, nil
}

func (e *Engine) RunCommand(ctx context.Context, cmd []string) (string, string, error) {
	newCtx := context.Background() // Create new context to avoid cancelling the parent context due to fetching new image/container

	containerID, err := e.GetOrCreateContainer(newCtx)
	if err != nil {
		log.Debug().Err(err).Msg("Error getting container")
		return "", "", err
	}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/app",
	}

	execID, err := e.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		log.Debug().Err(err).Msg("Error running command")
		return "", "", err
	}

	// Create response stream
	resp, err := e.client.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Error attaching to command")
		return "", "", err
	}
	defer resp.Close()

	// Read the output
	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
	if err != nil {
		return "", "", err
	}

	inspectResp, err := e.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return "", "", err
	}

	if inspectResp.ExitCode != 0 {
		return outBuf.String(), errBuf.String(),
			fmt.Errorf("command failed with exit code %d", inspectResp.ExitCode)
	}

	return outBuf.String(), errBuf.String(), nil
}

func (e *Engine) Cleanup(force bool) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.containerID != "" && (time.Since(e.lastUsed) > 10*time.Minute || force) {
		ctx := context.Background()
		err := e.client.ContainerStop(ctx, e.containerID, container.StopOptions{})
		if err != nil {
			return err
		}
		err = e.client.ContainerRemove(ctx, e.containerID, container.RemoveOptions{})
		if err != nil {
			return err
		}
		e.containerID = ""
	}
	return nil
}

func (e *Engine) isContainerHealthy(containerID string) bool {
	ctx := context.Background()
	inspect, err := e.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return false
	}
	return inspect.State.Running
}

func (e *Engine) LastUsed() time.Time {
	return e.lastUsed
}

func (e *Engine) ContainerID() string {
	return e.containerID
}

func (e *Engine) ContainerName() string {
	return e.containerName
}

func (e *Engine) Image() string {
	return e.image
}

func (e *Engine) GetCodePath() string {
	return e.CodePath
}
