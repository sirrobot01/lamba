package core

import (
	"context"
	"fmt"
	"time"
)

type Executor struct {
	FunctionRegistry *FunctionRegistry
	RuntimeManager   *RuntimeManager
}

func NewExecutor(registry *FunctionRegistry, manager *RuntimeManager) *Executor {
	return &Executor{
		FunctionRegistry: registry,
		RuntimeManager:   manager,
	}
}

func (e *Executor) Execute(name string, payload []byte) ([]byte, error) {
	metadata, exists := e.FunctionRegistry.Get(name)
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	runtime, exists := e.RuntimeManager.Get(metadata.Runtime)
	if !exists {
		return nil, fmt.Errorf("runtime %s not found", metadata.Runtime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return runtime.Execute(ctx, metadata, payload)
}

func (e *Executor) CreateFunction(metadata FunctionMetadata) error {
	runtime, exists := e.RuntimeManager.Get(metadata.Runtime)
	if !exists {
		return fmt.Errorf("runtime %s not found", metadata.Runtime)
	}

	if err := runtime.Prepare(metadata); err != nil {
		return err
	}

	e.FunctionRegistry.Register(metadata)
	return nil
}
