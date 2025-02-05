package containerd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultNamespace = "lamba"

type Engine struct {
	image         string
	containerID   string
	containerName string
	lastUsed      time.Time
	mutex         sync.Mutex
	CodePath      string
	client        *containerd.Client
}

func NewEngine(containerId, name, image, codePath string) *Engine {
	cl, err := containerd.New(
		"/run/containerd/containerd.sock",
		containerd.WithDefaultNamespace("lamba"),
		containerd.WithTimeout(60*time.Second),
	)
	if err != nil {
		panic(err)
	}
	// Validate image, append docker.io if not present
	if image == "" {
		panic("Image cannot be empty")
	}
	if !strings.Contains(image, "/") {
		image = "docker.io/library/" + image
	}
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
	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	_, err := e.client.GetImage(ctx, e.image)
	if err != nil {
		log.Info().Msgf("Pulling image: %s", e.image)
		_, err = e.client.Pull(ctx, e.image, containerd.WithPullUnpack)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) GetOrCreateContainer(ctx context.Context) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	// Check if container exists
	if e.containerID != "" {
		container, err := e.client.LoadContainer(ctx, e.containerID)
		if err == nil {
			task, err := container.Task(ctx, nil)
			if err == nil {
				status, err := task.Status(ctx)
				if err == nil && status.Status == containerd.Running {
					e.lastUsed = time.Now()
					return container.ID(), nil
				}
			}
		}
	}

	image, err := e.client.GetImage(ctx, e.image)
	if err != nil {
		log.Info().Msgf("Pulling image...")
		image, err = e.client.Pull(ctx, e.image, containerd.WithPullUnpack)
		if err != nil {
			return "", err
		}
	}

	// Create container
	container, err := e.client.NewContainer(
		ctx,
		e.containerName,
		containerd.WithNewSnapshot(e.containerName+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts([]specs.Mount{
				{
					Type:        "bind",
					Source:      e.CodePath,
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

	// Create and start the task
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return "", err
	}

	if err := task.Start(ctx); err != nil {
		return "", err
	}
	containerId := container.ID()
	e.containerID = containerId
	e.lastUsed = time.Now()
	return containerId, nil
}

func (e *Engine) RunCommand(ctx context.Context, cmd []string) (string, string, error) {

	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	containerId, err := e.GetOrCreateContainer(ctx)
	if err != nil || containerId == "" {
		return "", "", err
	}
	container, err := e.client.LoadContainer(ctx, containerId)
	if err != nil {
		return "", "", err
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "", "", err
	}

	// Create and configure the process
	spec, err := container.Spec(ctx)
	if err != nil {
		return "", "", err
	}

	spec.Process.Args = cmd
	spec.Process.Cwd = "/app"

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	process, err := task.Exec(ctx,
		fmt.Sprintf("%s-exec-%d", e.containerName, time.Now().UnixNano()),
		spec.Process,
		cio.NewCreator(
			cio.WithStreams(nil, stdout, stderr),
		),
	)
	if err != nil {
		return "", "", err
	}

	// Start the process
	if err := process.Start(ctx); err != nil {
		return "", "", err
	}

	// Wait for the process to complete
	exitCh, err := process.Wait(ctx)
	if err != nil {
		return "", "", err
	}

	select {
	case status := <-exitCh:
		// Clean up the process
		defer func(process containerd.Process, ctx context.Context, opts ...containerd.ProcessDeleteOpts) {
			_, err := process.Delete(ctx, opts...)
			if err != nil {
				log.Info().Msgf("Failed to delete process: %v\n", err)
			}
		}(process, ctx)

		if status.ExitCode() != 0 {
			return stdout.String(), stderr.String(),
				fmt.Errorf("command failed with exit code %d: %s", status.ExitCode(), stderr.String())
		}
	case <-ctx.Done():
		return "", "", ctx.Err()
	}

	return stdout.String(), stderr.String(), nil
}

func (e *Engine) Cleanup(force bool) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	ctx := context.Background()
	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	if e.containerID != "" && (time.Since(e.lastUsed) > 10*time.Minute || force) {

		container, err := e.client.LoadContainer(ctx, e.containerID)
		if err != nil {
			return err
		}

		task, err := container.Task(ctx, nil)
		if err != nil {
			return err
		}

		// Kill the task
		if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
			return err
		}

		// Wait for the task to exit
		_, err = task.Wait(ctx)
		if err != nil {
			return err
		}

		// Delete the task
		if _, err := task.Delete(ctx); err != nil {
			return err
		}

		// Delete the container
		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			return err
		}

		e.containerID = ""
	}
	return nil
}

func (e *Engine) isContainerHealthy(containerID string) bool {
	ctx := context.Background()
	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	container, err := e.client.LoadContainer(ctx, containerID)
	if err != nil {
		return false
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return false
	}

	status, err := task.Status(ctx)
	if err != nil {
		return false
	}

	return status.Status == containerd.Running
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
