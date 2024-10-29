package runtime

import (
	"bytes"
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
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

	if cm.containerID != "" {
		container, err := containerdClient.LoadContainer(ctx, cm.containerID)
		if err == nil {
			task, err := container.Task(ctx, nil)
			if err == nil {
				status, err := task.Status(ctx)
				if err == nil && status.Status == containerd.Running {
					cm.lastUsed = time.Now()
					return cm.containerID, nil
				}
			}
		}
	}

	image, err := containerdClient.GetImage(ctx, cm.image)
	if err != nil {
		return "", err
	}

	container, err := containerdClient.NewContainer(
		ctx,
		cm.containerName,
		containerd.WithNewSnapshot(cm.containerName+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithHostNamespace(specs.NetworkNamespace),
			oci.WithMounts([]specs.Mount{
				{
					Type:        "bind",
					Source:      cm.CodePath,
					Destination: "/app",
					Options:     []string{"rbind", "rw"},
				},
			}),
			oci.WithProcessArgs("tail", "-f", "/dev/null"),
		),
	)
	if err != nil {
		return "", err
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return "", err
	}

	if err := task.Start(ctx); err != nil {
		return "", err
	}

	cm.containerID = container.ID()
	cm.lastUsed = time.Now()
	return container.ID(), nil
}

func (cm *ContainerManager) RunCommand(ctx context.Context, cmd []string) (string, string, error) {
	containerID, err := cm.getOrCreateContainer(ctx)
	if err != nil {
		return "", "", err
	}

	container, err := containerdClient.LoadContainer(ctx, containerID)
	if err != nil {
		return "", "", err
	}

	_, err = container.Spec(ctx)
	if err != nil {
		return "", "", err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "", "", err
	}

	var stdout, stderr bytes.Buffer
	process, err := task.Exec(ctx,
		cm.containerName+"-exec",
		&specs.Process{
			Args: cmd,
			Cwd:  "/app",
		},
		cio.NewCreator(
			cio.WithStreams(nil, &stdout, &stderr),
		),
	)
	if err != nil {
		return "", "", err
	}

	statusC, err := process.Wait(ctx)
	if err != nil {
		return "", "", err
	}

	if err := process.Start(ctx); err != nil {
		return "", "", err
	}

	status := <-statusC
	code, _, err := status.Result()
	if err != nil {
		return "", "", err
	}

	if code != 0 {
		return stdout.String(), stderr.String(),
			fmt.Errorf("command failed with exit code %d", code)
	}

	return stdout.String(), stderr.String(), nil
}

func (cm *ContainerManager) Cleanup(force bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.containerID != "" && (time.Since(cm.lastUsed) > 10*time.Minute || force) {
		ctx := context.Background()
		container, err := containerdClient.LoadContainer(ctx, cm.containerID)
		if err != nil {
			return err
		}

		task, err := container.Task(ctx, nil)
		if err == nil {
			_, err = task.Delete(ctx, containerd.WithProcessKill)
			if err != nil {
				return err
			}
		}

		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			return err
		}

		cm.containerID = ""
	}
	return nil
}
