package executor

import (
	"context"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"time"
)

type Executor struct {
	FunctionRegistry *function.Registry
	RuntimeManager   *runtime.Manager
	EventManager     *event.Manager
}

func NewExecutor(registry *function.Registry, rtnManager *runtime.Manager, eventManager *event.Manager) *Executor {
	return &Executor{
		FunctionRegistry: registry,
		RuntimeManager:   rtnManager,
		EventManager:     eventManager,
	}
}

func (e *Executor) Execute(invoker, funcName string, payload []byte) ([]byte, error) {
	metadata, exists := e.FunctionRegistry.Get(funcName)
	if !exists {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	rtn, exists := e.RuntimeManager.Get(metadata.Runtime)
	if !exists {
		return nil, fmt.Errorf("runtime %s not found", metadata.Runtime)
	}
	ev := e.EventManager.Add(invoker, funcName, metadata.Runtime, payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(metadata.Timeout)*time.Second)
	defer cancel()
	result, err := runtime.Execute(rtn, ctx, ev, metadata)
	ev.Result = result
	if err != nil {
		e.EventManager.MarkFailed(ev, err)
		return nil, err
	}
	e.EventManager.MarkCompleted(ev)
	return result, nil
}

func (e *Executor) CreateFunction(name string, runtime string, handler string, timeout int, codePath string, preExec string) error {
	fn := function.Function{
		Name:     name,
		Runtime:  runtime,
		Handler:  handler,
		CodePath: codePath,
		Timeout:  timeout,
		PreExec:  preExec,
	}
	rtn, exists := e.RuntimeManager.Get(fn.Runtime)
	if !exists {
		return fmt.Errorf("runtime %s not found", fn.Runtime)
	}

	if err := rtn.Init(); err != nil {
		return err
	}

	e.FunctionRegistry.Register(fn)
	return nil
}

func (e *Executor) DeleteFunction(name string) error {
	e.FunctionRegistry.Delete(name)
	return nil
}
