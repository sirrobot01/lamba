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
	"github.com/sirrobot01/lamba/internal/function"
	"strings"
	"sync"
	"syscall"
	"time"
)

const defaultNamespace = "lamba"

type Engine struct {
	image    string
	lastUsed time.Time
	mutex    sync.Mutex
	client   *containerd.Client
}

func NewEngine(image string) *Engine {
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
		client: cl,
		image:  image,
	}
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

func (e *Engine) GetOrCreateContainer(ctx context.Context, fn *function.Function) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	// Check if container exists
	if fn.ContainerID != "" {
		container, err := e.client.LoadContainer(ctx, fn.ContainerID)
		if err == nil {
			task, err := container.Task(ctx, nil)
			if err == nil {
				status, err := task.Status(ctx)
				if err == nil && status.Status == containerd.Running {
					e.lastUsed = time.Now()
					return nil
				}
			}
		}
	}

	image, err := e.client.GetImage(ctx, e.image)
	if err != nil {
		log.Info().Msgf("Pulling image...")
		image, err = e.client.Pull(ctx, e.image, containerd.WithPullUnpack)
		if err != nil {
			return err
		}
	}

	// Create container
	container, err := e.client.NewContainer(
		ctx,
		fn.ContainerName,
		containerd.WithNewSnapshot(fn.ContainerName+"snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts([]specs.Mount{
				{
					Type:        "bind",
					Source:      fn.CodePath,
					Destination: "/app",
					Options:     []string{"rbind", "rw"},
				},
			}),
			oci.WithProcessArgs("tail", "-f", "/dev/null"),
		),
	)
	if err != nil {
		return err
	}

	// Create and start the task
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}

	if err := task.Start(ctx); err != nil {
		return err
	}
	fn.ContainerID = container.ID()
	e.lastUsed = time.Now()
	return nil
}

func (e *Engine) RunCommand(ctx context.Context, fn *function.Function, cmd []string) (string, string, error) {

	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	if err := e.GetOrCreateContainer(ctx, fn); err != nil || fn.ContainerID == "" {
		return "", "", err
	}
	container, err := e.client.LoadContainer(ctx, fn.ContainerID)
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
		fmt.Sprintf("%s-exec-%d", fn.ContainerName, time.Now().UnixNano()),
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

func (e *Engine) Cleanup(fn *function.Function, force bool) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	ctx := context.Background()
	ctx = namespaces.WithNamespace(ctx, defaultNamespace)

	if fn.ContainerID != "" && (time.Since(e.lastUsed) > 10*time.Minute || force) {

		container, err := e.client.LoadContainer(ctx, fn.ContainerID)
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

		fn.ContainerID = ""
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
