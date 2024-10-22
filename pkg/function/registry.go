package function

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Registry struct {
	functions map[string]Function
	mu        sync.RWMutex
	filePath  string
}

func NewRegistry() *Registry {
	currentDir, _ := os.Getwd()
	filePath := filepath.Join(currentDir, ".lamba_functions.json")
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
			fmt.Printf("Error reading function data: %v\n", err)
		}
		return
	}

	err = json.Unmarshal(data, &fr.functions)
	if err != nil {
		fmt.Printf("Error parsing function data: %v\n", err)
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

func (fr *Registry) Register(metadata Function) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.functions[metadata.Name] = metadata
	_ = fr.saveToFile()
}

func (fr *Registry) Get(name string) (Function, bool) {
	metadata, exists := fr.functions[name]
	return metadata, exists
}

func (fr *Registry) List() []Function {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	functions := make([]Function, 0, len(fr.functions))
	for _, metadata := range fr.functions {
		functions = append(functions, metadata)
	}
	return functions
}
