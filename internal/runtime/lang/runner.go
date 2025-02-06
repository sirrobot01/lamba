package lang

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/internal/engines"
	"github.com/sirrobot01/lamba/internal/engines/containerd"
	"github.com/sirrobot01/lamba/internal/engines/docker"
	"github.com/sirrobot01/lamba/internal/event"
	"github.com/sirrobot01/lamba/internal/function"
	"time"
)

type Runtime struct {
	Engine  engines.Engine
	name    string
	image   string
	version string
}

func NewRuntime(engine engines.Type, name, image, version string) Runtime {
	var e engines.Engine
	switch engine {
	case engines.Docker:
		e = docker.NewEngine(image)
	case engines.Containerd:
		e = containerd.NewEngine(image)
	default:
		e = docker.NewEngine(image)
	}
	return Runtime{
		Engine:  e,
		name:    name,
		image:   image,
		version: version,
	}
}

func (r *Runtime) Init(fn *function.Function) error {
	var err error

	ctx := context.Background()

	// Check if image exists, pull if not
	if err = r.Engine.PullImage(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to pull image")
		return err
	}

	if fn != nil {
		if err := r.Engine.GetOrCreateContainer(ctx, fn); err != nil {
			return err
		}
		if fn.ContainerID == "" {
			return fmt.Errorf("container not found")
		}

	}
	return nil
}

func (r *Runtime) Shutdown(fn *function.Function) error {
	// Stop and remove container, gracefully
	return r.Engine.Cleanup(fn, true)
}

func (r *Runtime) GetImage() string {
	return r.image
}

func (r *Runtime) GetCmd(event *event.Event, fn *function.Function) []string {
	switch r.name {
	case "python":
		return r.GetPythonCmd(event, fn)
	case "nodejs":
		return r.GetNodeJSCmd(event, fn)
	case "golang":
		return r.GetGoCmd(event, fn)
	default:
		cmd := []string{fn.Handler}
		payload := event.GetPayload()
		if payload != "" {
			cmd = append(cmd, payload)
		}
		return cmd
	}
}

func (r *Runtime) GetEngine() engines.Engine {
	return r.Engine
}

func (r *Runtime) Run(ctx context.Context, event *event.Event, fn *function.Function) (string, error) {
	cmd := r.GetCmd(event, fn)

	stdout, stderr, err := r.Engine.RunCommand(ctx, fn, cmd)

	fn.LastRun = time.Now()

	if err != nil || stderr != "" {
		return "", fmt.Errorf("error: %s, stderr: %s", err, stderr)
	}
	return stdout, nil
}
