package core

import (
	"context"
)

type RuntimeManager struct {
	runtimes map[string]Runtime
}

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		runtimes: make(map[string]Runtime),
	}
}

func (m *RuntimeManager) Register(name string, runtime Runtime) {
	m.runtimes[name] = runtime
}

func (m *RuntimeManager) Get(name string) (Runtime, bool) {
	runtime, exists := m.runtimes[name]
	return runtime, exists
}

func (m *RuntimeManager) List() []string {
	runtimes := make([]string, 0, len(m.runtimes))
	for runtime := range m.runtimes {
		runtimes = append(runtimes, runtime)
	}
	return runtimes
}

func (m *RuntimeManager) Shutdown() {
	for _, runtime := range m.runtimes {
		err := runtime.Shutdown()
		if err != nil {
			return
		}
	}
}

type Runtime interface {
	Execute(ctx context.Context, metadata FunctionMetadata, payload []byte) ([]byte, error)
	Init() error
	Shutdown() error
	GetName() string
	GetVersion() string
	Validate(metadata FunctionMetadata) error
	Prepare(metadata FunctionMetadata) error
}
