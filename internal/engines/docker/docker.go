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
	"github.com/sirrobot01/lamba/internal/function"
	"io"
	"sync"
	"time"
)

type Engine struct {
	image    string
	lastUsed time.Time
	mutex    sync.Mutex
	CodePath string
	client   *client.Client
}

func NewEngine(image string) *Engine {
	cl, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &Engine{
		client: cl,
		image:  image,
	}
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

func (e *Engine) GetOrCreateContainer(ctx context.Context, fn *function.Function) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if container exists and is recent
	if fn.ContainerID != "" {
		// Check if container is still running
		_, err := e.client.ContainerInspect(ctx, fn.ContainerID)
		if err == nil {
			e.lastUsed = time.Now()
			return nil
		}
	}

	// Check if image exists, pull if not

	if err := e.PullImage(ctx); err != nil {
		return err
	}

	// Create new container
	hostConf := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app", fn.CodePath),
		},
	}
	resp, err := e.client.ContainerCreate(ctx,
		&container.Config{
			Image:      e.image,
			Cmd:        []string{"tail", "-f", "/dev/null"}, // Keep container running
			Tty:        true,
			WorkingDir: "/app",
		},
		hostConf, nil, nil, fn.ContainerName)
	if err != nil {
		return err
	}
	// Start container
	if err := e.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	fn.ContainerID = resp.ID
	e.lastUsed = time.Now()
	return nil
}

func (e *Engine) RunCommand(ctx context.Context, fn *function.Function, cmd []string) (string, string, error) {
	newCtx := context.Background() // Create new context to avoid cancelling the parent context due to fetching new image/container

	if err := e.GetOrCreateContainer(newCtx, fn); err != nil {
		log.Debug().Err(err).Msg("Error getting container")
		return "", "", err
	}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/app",
	}

	execID, err := e.client.ContainerExecCreate(ctx, fn.ContainerID, execConfig)
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

func (e *Engine) Cleanup(fn *function.Function, force bool) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if fn.ContainerID != "" && (time.Since(e.lastUsed) > 10*time.Minute || force) {
		ctx := context.Background()
		err := e.client.ContainerStop(ctx, fn.ContainerID, container.StopOptions{})
		if err != nil {
			return err
		}
		err = e.client.ContainerRemove(ctx, fn.ContainerID, container.RemoveOptions{})
		if err != nil {
			return err
		}
		fn.ContainerID = ""
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
