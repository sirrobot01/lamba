package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/common"
	"github.com/sirrobot01/lamba/internal/event"
	"github.com/sirrobot01/lamba/internal/function"
	"github.com/sirrobot01/lamba/internal/runtime"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Executor struct {
	FunctionRegistry *function.Registry
	RuntimeManager   *runtime.Manager
	EventManager     *event.Manager
	MemoryManager    *runtime.MemoryManager
}

func New(registry *function.Registry, rtnManager *runtime.Manager, eventManager *event.Manager, memory *runtime.MemoryManager) *Executor {
	return &Executor{
		FunctionRegistry: registry,
		RuntimeManager:   rtnManager,
		EventManager:     eventManager,
		MemoryManager:    memory,
	}
}

func (e *Executor) Execute(invoker, functionId string, payload string) (string, error) {
	if functionId == "" {
		return "", fmt.Errorf("invalid functionID")
	}

	fn, exists := e.FunctionRegistry.Get(functionId)
	if !exists {
		return "", fmt.Errorf("function %s not found", functionId)
	}

	log.Debug().Msgf("[%s] Executing function %s with payload %s", invoker, fn.Name, payload)

	rtn, exists := e.RuntimeManager.Get(fn.Runtime)
	if !exists {
		return "", fmt.Errorf("runtime %s not found", fn.Runtime)
	}
	ev := e.EventManager.Add(invoker, fn, payload)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fn.Timeout)*time.Second)
	defer cancel()
	result, err := rtn.Run(ctx, ev, fn)
	if err != nil {
		log.Info().Err(err).Msgf("Error executing function %s", fn.Name)
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
	e.FunctionRegistry.Update(*fn)
	return result, nil
}

func (e *Executor) CreateFunctionDir(file io.ReaderAt, name, runtime string, size int64) (string, error) {
	rtnPath := filepath.Join("assets/functions", runtime)
	if err := os.MkdirAll(rtnPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}
	functionDir, _ := filepath.Abs(filepath.Join(rtnPath, name))
	if err := common.ExtractZip(file, size, rtnPath); err != nil {
		err := os.RemoveAll(functionDir)
		return "", fmt.Errorf("error extrating file: %v", err)
	}
	return functionDir, nil
}

func (e *Executor) CreateFunction(name string, runtime string, handler string, timeout int, preExec string, file io.ReaderAt, fileSize int64) error {

	codePath, err := e.CreateFunctionDir(file, name, runtime, fileSize)
	if err != nil {
		return err
	}

	fn := function.Function{
		ID:            uuid.New().String(),
		Name:          name,
		Runtime:       runtime,
		Handler:       handler,
		CodePath:      codePath,
		ContainerName: fmt.Sprintf("%s-%s", name, runtime),
		Timeout:       timeout,
		PreExec:       preExec,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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

func (e *Executor) DeleteFunction(id string) error {
	// Get function
	fn, exists := e.FunctionRegistry.Get(id)
	if !exists {
		return fmt.Errorf("function %s not found", id)
	}
	go func(f *function.Function) {
		// Get runtime
		rtn, ok := e.RuntimeManager.Get(f.Runtime)
		// Shutdown Container
		if ok {
			if err := rtn.Shutdown(f); err != nil {
				log.Info().Err(err).Msgf("Error shutting down function %s", f.Name)
			}
		} else {
			log.Info().Msgf("Runtime %s not found", f.Runtime)
		}
	}(fn)

	// Delete function
	e.FunctionRegistry.Delete(fn)

	return nil
}
