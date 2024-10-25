package runtime

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirrobot01/lamba/pkg/function"
	"sync"
	"time"
)

type ContainerManager struct {
	image         string
	containerID   string
	containerName string
	lastUsed      time.Time
	mutex         sync.Mutex
	CodePath      string
}

func NewContainerManager(fn *function.Function, image string) *ContainerManager {
	return &ContainerManager{
		image:         image,
		containerID:   fn.ContainerID,
		containerName: fn.Name,
		lastUsed:      fn.LastRun,
		CodePath:      fn.CodePath,
	}
}

func (cm *ContainerManager) getOrCreateContainer(ctx context.Context) (string, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if container exists and is recent
	if cm.containerID != "" {
		// Check if container is still running
		_, err := dockerClient.ContainerInspect(ctx, cm.containerID)
		if err == nil {
			cm.lastUsed = time.Now()
			return cm.containerID, nil
		}
	}

	// Create new container
	hostConf := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app", cm.CodePath),
		},
	}
	resp, err := dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:      cm.image,
			Cmd:        []string{"tail", "-f", "/dev/null"}, // Keep container running
			Tty:        true,
			WorkingDir: "/app",
		},
		hostConf, nil, nil, cm.containerName)
	if err != nil {
		return "", err
	}
	// Start container
	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	cm.containerID = resp.ID
	cm.lastUsed = time.Now()
	return resp.ID, nil
}

func (cm *ContainerManager) RunCommand(ctx context.Context, cmd []string) (string, string, error) {
	newCtx := context.Background() // Create new context to avoid cancelling the parent context due to fetching new image/container
	containerID, err := cm.getOrCreateContainer(newCtx)
	if err != nil {
		return "", "", err
	}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/app",
	}

	execID, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", "", err
	}

	// Create response stream
	resp, err := dockerClient.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", "", err
	}
	defer resp.Close()

	// Read the output
	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
	if err != nil {
		return "", "", err
	}

	inspectResp, err := dockerClient.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return "", "", err
	}

	if inspectResp.ExitCode != 0 {
		return outBuf.String(), errBuf.String(),
			fmt.Errorf("command failed with exit code %d", inspectResp.ExitCode)
	}

	return outBuf.String(), errBuf.String(), nil
}

func (cm *ContainerManager) Cleanup(force bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.containerID != "" && (time.Since(cm.lastUsed) > 10*time.Minute || force) {
		ctx := context.Background()
		err := dockerClient.ContainerStop(ctx, cm.containerID, container.StopOptions{})
		if err != nil {
			return err
		}
		err = dockerClient.ContainerRemove(ctx, cm.containerID, container.RemoveOptions{})
		if err != nil {
			return err
		}
		cm.containerID = ""
	}
	return nil
}

func (cm *ContainerManager) isContainerHealthy(containerID string) bool {
	ctx := context.Background()
	inspect, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return false
	}
	return inspect.State.Running
}
