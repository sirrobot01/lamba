package executor

import (
	"context"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	function2 "github.com/sirrobot01/lamba/pkg/function"
	"github.com/sirrobot01/lamba/pkg/runtime"
	"time"
)

type Executor struct {
	FunctionRegistry *function2.Registry
	RuntimeManager   *runtime.Manager
}

func NewExecutor(registry *function2.Registry, manager *runtime.Manager) *Executor {
	return &Executor{
		FunctionRegistry: registry,
		RuntimeManager:   manager,
	}
}

func (e *Executor) Execute(invoker, funcName string, payload []byte) ([]byte, error) {
	ev := event.InvokeEvent{
		Name:      invoker,
		Payload:   payload,
		InvokedAt: time.Now(),
	}
	metadata, exists := e.FunctionRegistry.Get(funcName)
	if !exists {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	rtn, exists := e.RuntimeManager.Get(metadata.Runtime)
	if !exists {
		return nil, fmt.Errorf("runtime %s not found", metadata.Runtime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(metadata.Timeout)*time.Second)
	defer cancel()
	return rtn.Execute(ctx, ev, metadata)
}

func (e *Executor) CreateFunction(name string, runtime string, handler string, timeout int, codePath string, preExec string) error {
	fn := function2.Function{
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
