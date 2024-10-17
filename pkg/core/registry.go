package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FunctionMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Runtime     string `json:"runtime"`
	Handler     string `json:"handler"`
	CodePath    string `json:"codePath"`
	Timeout     int    `json:"timeout"`
}

func NewFunctionMetadata(name string, runtime string, handler string, timeout int, codePath string) FunctionMetadata {
	return FunctionMetadata{
		Name:     name,
		Runtime:  runtime,
		Handler:  handler,
		CodePath: codePath,
		Timeout:  timeout,
	}
}

type FunctionRegistry struct {
	functions map[string]FunctionMetadata
	mu        sync.RWMutex
	filePath  string
}

func NewFunctionRegistry() *FunctionRegistry {
	currentDir, _ := os.Getwd()
	filePath := filepath.Join(currentDir, ".lamba_functions.json")
	fr := &FunctionRegistry{
		functions: make(map[string]FunctionMetadata),
		filePath:  filePath,
	}
	fr.loadFromFile()
	return fr
}

func (fr *FunctionRegistry) loadFromFile() {
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
func (fr *FunctionRegistry) saveToFile() error {
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

func (fr *FunctionRegistry) Register(metadata FunctionMetadata) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.functions[metadata.Name] = metadata
	_ = fr.saveToFile()
}

func (fr *FunctionRegistry) Get(name string) (FunctionMetadata, bool) {
	metadata, exists := fr.functions[name]
	return metadata, exists
}

func (fr *FunctionRegistry) List() []FunctionMetadata {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	functions := make([]FunctionMetadata, 0, len(fr.functions))
	for _, metadata := range fr.functions {
		functions = append(functions, metadata)
	}
	return functions
}
