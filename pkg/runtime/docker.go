package runtime

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

type ContainerdRuntime struct {
	name    string
	image   string
	version string
}

func NewContainerdRuntime(name, image, version string) (*ContainerdRuntime, error) {

	return &ContainerdRuntime{
		name:    name,
		image:   image,
		version: version,
	}, nil
}

func (cr *ContainerdRuntime) Init(fn *function.Function) error {
	ctx := namespaces.WithNamespace(context.Background(), "lamba")

	// Pull image if it doesn't exist
	_, err := containerdClient.GetImage(ctx, cr.image)
	if err != nil {
		fmt.Println("Pulling image...")
		_, err = containerdClient.Pull(ctx, cr.image, containerd.WithPullUnpack)
		if err != nil {
			return err
		}
	}

	if fn != nil {
		containerId, err := cr.createContainer(ctx, fn)
		if err != nil {
			return err
		}
		fn.ContainerID = containerId
	}
	return nil
}

func (cr *ContainerdRuntime) createContainer(ctx context.Context, fn *function.Function) (string, error) {
	cm := NewContainerManager(fn, cr.GetImage())
	return cm.getOrCreateContainer(ctx)
}

func (cr *ContainerdRuntime) Shutdown(fn *function.Function) error {
	cm := NewContainerManager(fn, cr.GetImage())
	return cm.Cleanup(true)
}

func (cr *ContainerdRuntime) GetImage() string {
	return cr.image
}

func (cr *ContainerdRuntime) GetCmd(event event.Event, fn function.Function) []string {
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
