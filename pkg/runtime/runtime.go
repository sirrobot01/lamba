package runtime

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"sync"
)

var dockerClient *client.Client

func GetDockerClient() *client.Client {
	return dockerClient
}

type Runtime interface {
	Init() error
	Shutdown() error
	GetCmd(event event.InvokeEvent, fn function.Function) []string
	Execute(ctx context.Context, event event.InvokeEvent, fn function.Function) ([]byte, error)
}

type Manager struct {
	runtimes map[string]Runtime
}

func NewManager() *Manager {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	runtimes := make(map[string]Runtime)
	return &Manager{
		runtimes: runtimes,
	}
}

func (m *Manager) Register(runtimes map[string]Runtime) {
	var wg sync.WaitGroup
	for name, runtime := range runtimes {
		wg.Add(1)
		go func(name string, runtime Runtime) {
			defer wg.Done()
			err := runtime.Init()
			if err != nil {
				panic(err)
			}
			m.runtimes[name] = runtime
		}(name, runtime)
	}
	wg.Wait()

}

func (m *Manager) Get(name string) (Runtime, bool) {
	runtime, exists := m.runtimes[name]
	return runtime, exists
}

func (m *Manager) List() []string {
	runtimes := make([]string, 0, len(m.runtimes))
	for runtime := range m.runtimes {
		runtimes = append(runtimes, runtime)
	}
	return runtimes
}

func (m *Manager) Shutdown() {
	for _, runtime := range m.runtimes {
		err := runtime.Shutdown()
		if err != nil {
			return
		}
	}
}
