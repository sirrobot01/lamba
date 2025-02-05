package languages

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime/engines"
	"github.com/sirrobot01/lamba/pkg/runtime/engines/containerd"
	"github.com/sirrobot01/lamba/pkg/runtime/engines/docker"
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
		e = docker.NewEngine("", name, image, "")
	case engines.Containerd:
		e = containerd.NewEngine("", name, image, "")
	default:
		e = docker.NewEngine("", name, image, "")
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
		r.Engine.UpdateEngine("", fn.Name, fn.CodePath)
		containerId, err := r.Engine.GetOrCreateContainer(ctx)
		if err != nil {
			return err
		}
		if containerId == "" {
			return fmt.Errorf("container not found")
		}
		fn.ContainerID = containerId
		r.Engine.UpdateEngine(containerId, fn.Name, fn.CodePath)

	}
	return nil
}

func (r *Runtime) Shutdown(fn *function.Function) error {
	// Stop and remove container, gracefully
	return r.Engine.Cleanup(true)
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
	case "go":
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

	r.Engine.UpdateEngine(fn.ContainerID, fn.Name, fn.CodePath)

	stdout, stderr, err := r.Engine.RunCommand(ctx, cmd)

	fn.LastRun = r.Engine.LastUsed()
	fn.ContainerID = r.Engine.ContainerID()

	if err != nil || stderr != "" {
		return "", fmt.Errorf("error: %s, stderr: %s", err, stderr)
	}
	return stdout, nil
}
