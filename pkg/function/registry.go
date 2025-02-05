package function

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sync"
)

type Registry struct {
	functions map[string]Function
	mu        sync.RWMutex
	filePath  string
}

func NewRegistry(configDir string) *Registry {
	filePath := filepath.Join(configDir, "db", "functions.json")
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil
	}
	fr := &Registry{
		functions: make(map[string]Function),
		filePath:  filePath,
	}
	fr.loadFromFile()
	return fr
}

func (fr *Registry) loadFromFile() {
	// Load the functions from the file
	data, err := os.ReadFile(fr.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Info().Msgf("Error reading function data: %v\n", err)
		}
		return
	}

	err = json.Unmarshal(data, &fr.functions)
	if err != nil {
		log.Info().Msgf("Error parsing function data: %v\n", err)
	}
}
func (fr *Registry) saveToFile() error {
	data, err := json.Marshal(fr.functions)
	if err != nil {
		return fmt.Errorf("error encoding function data: %v", err)
	}

	err = os.WriteFile(fr.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing function data: %v", err)
	}

	return nil
}

func (fr *Registry) Register(fn Function) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.functions[fn.Name] = fn
	go func() {
		_ = fr.saveToFile()
	}()
}

func (fr *Registry) Get(name string) (Function, bool) {
	fn, exists := fr.functions[name]
	return fn, exists
}

func (fr *Registry) List() []Function {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	functions := make([]Function, 0, len(fr.functions))
	for _, fn := range fr.functions {
		functions = append(functions, fn)
	}
	return functions
}

func (fr *Registry) Delete(name string) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	delete(fr.functions, name)
	go func() {
		_ = fr.saveToFile()
	}()
}

func (fr *Registry) Update(fn Function) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.functions[fn.Name] = fn
	go func() {
		_ = fr.saveToFile()
	}()
}
