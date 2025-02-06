package runtime

import (
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/internal/engines"
	"github.com/sirrobot01/lamba/internal/function"
	"github.com/sirrobot01/lamba/internal/runtime/lang"
	"sync"
)

type Manager struct {
	engine   engines.Type
	runtimes map[string]lang.Runtime
	mu       sync.Mutex
}

func NewManager(engine engines.Type) *Manager {
	return &Manager{
		engine:   engine,
		runtimes: make(map[string]lang.Runtime),
	}
}

func (m *Manager) Register(runtimes map[string]lang.Runtime) error {
	log.Info().Msgf("Registering runtimes on %s engine", m.engine)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(runtimes))

	for name, runtime := range runtimes {
		log.Debug().Msgf("Registering runtime: %s", name)
		wg.Add(1)
		go func(name string, runtime lang.Runtime) {
			defer wg.Done()
			if err := runtime.Init(nil); err != nil {
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

func (m *Manager) Get(name string) (lang.Runtime, bool) {
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

func (m *Manager) Shutdown(fn *function.Function) {
	for _, runtime := range m.runtimes {
		err := runtime.Shutdown(fn)
		if err != nil {
			return
		}
	}
}
