package runtime

import (
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
	GetCmd(event event.Event, fn function.Function) []string
	GetImage() string
}

type Manager struct {
	runtimes map[string]Runtime
	mu       sync.Mutex
}

func NewManager() *Manager {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Manager{
		runtimes: make(map[string]Runtime),
	}
}

func (m *Manager) Register(runtimes map[string]Runtime) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(runtimes))

	for name, runtime := range runtimes {
		wg.Add(1)
		go func(name string, runtime Runtime) {
			defer wg.Done()
			if err := runtime.Init(); err != nil {
				errorChan <- err
			}
			m.mu.Lock()
			m.runtimes[name] = runtime
			m.mu.Unlock()
		}(name, runtime)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	// Close the channel after all goroutines are done
	close(errorChan)

	// Check for any errors
	for err := range errorChan {
		if err != nil {
			return err // Return the first error encountered
		}
	}

	return nil

}

func (m *Manager) Get(name string) (Runtime, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
