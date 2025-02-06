package engines

import (
	"context"
	"fmt"
	"github.com/sirrobot01/lamba/internal/function"
)

// Type is the type of engine that the runtime uses
type Type string

const (
	// Docker engine
	Docker Type = "docker"
	// Containerd engine
	Containerd Type = "containerd"
)

func (t *Type) String() string {
	return string(*t)
}

func (t *Type) Set(value string) error {
	switch Type(value) {
	case Docker, Containerd:
		*t = Type(value)
		return nil
	default:
		return fmt.Errorf("invalid engine type: %s", value)
	}
}

// Engine is the interface that wraps the basic methods of an engine
type Engine interface {
	GetOrCreateContainer(ctx context.Context, fn *function.Function) error
	RunCommand(ctx context.Context, fn *function.Function, cmd []string) (string, string, error)
	Cleanup(fn *function.Function, force bool) error
	PullImage(ctx context.Context) error
}
