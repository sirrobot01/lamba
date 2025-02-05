package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"time"
)

type Executor struct {
	FunctionRegistry *function.Registry
	RuntimeManager   *runtime.Manager
	EventManager     *event.Manager
	MemoryManager    *runtime.MemoryManager
}

func NewExecutor(registry *function.Registry, rtnManager *runtime.Manager, eventManager *event.Manager, memory *runtime.MemoryManager) *Executor {
	return &Executor{
		FunctionRegistry: registry,
		RuntimeManager:   rtnManager,
		EventManager:     eventManager,
		MemoryManager:    memory,
	}
}

func (e *Executor) Execute(invoker, funcName string, payload string) (string, error) {
	log.Debug().Msgf("[%s] Executing function %s with payload %s", invoker, funcName, payload)
	fn, exists := e.FunctionRegistry.Get(funcName)
	if !exists {
		return "", fmt.Errorf("function %s not found", funcName)
	}

	rtn, exists := e.RuntimeManager.Get(fn.Runtime)
	if !exists {
		return "", fmt.Errorf("runtime %s not found", fn.Runtime)
	}
	ev := e.EventManager.Add(invoker, funcName, fn.Runtime, payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fn.Timeout)*time.Second)
	defer cancel()
	result, err := rtn.Run(ctx, ev, &fn)
	if err != nil {
		log.Info().Err(err).Msgf("Error executing function %s", funcName)
		e.EventManager.MarkFailed(ev, err)
		return "", err
	}
	var res struct {
		Result any      `json:"result"`
		Debug  []string `json:"debug"`
	}

	if err := json.Unmarshal([]byte(result), &res); err != nil {
		return "", err
	}
	ev.Result = res.Result
	ev.Debug = res.Debug
	e.EventManager.MarkCompleted(ev)
	e.FunctionRegistry.Update(fn)
	return result, nil
}

func (e *Executor) CreateFunction(name string, runtime string, handler string, timeout int, codePath string, preExec string) error {
	fn := function.Function{
		Name:      name,
		Runtime:   runtime,
		Handler:   handler,
		CodePath:  codePath,
		Timeout:   timeout,
		PreExec:   preExec,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	rtn, exists := e.RuntimeManager.Get(fn.Runtime)
	if !exists {
		return fmt.Errorf("runtime %s not found", fn.Runtime)
	}

	if err := rtn.Init(&fn); err != nil {
		return err
	}

	e.FunctionRegistry.Register(fn)
	return nil
}

func (e *Executor) DeleteFunction(name string) error {
	// Get function
	fn, exists := e.FunctionRegistry.Get(name)
	if !exists {
		return fmt.Errorf("function %s not found", name)
	}
	go func(f function.Function) {
		// Get runtime
		rtn, ok := e.RuntimeManager.Get(f.Runtime)
		// Shutdown Container
		if ok {
			if err := rtn.Shutdown(&f); err != nil {
				log.Info().Err(err).Msgf("Error shutting down function %s", f.Name)
			}
		} else {
			log.Info().Msgf("Runtime %s not found", f.Runtime)
		}
	}(fn)

	// Delete function
	e.FunctionRegistry.Delete(name)

	return nil
}
