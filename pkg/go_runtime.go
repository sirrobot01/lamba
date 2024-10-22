package pkg

//
//import (
//	"context"
//	"fmt"
//	"github.com/sirrobot01/lamba/common"
//	"plugin"
//	"sync"
//)
//
//type GoRuntime struct {
//	functions map[string]*plugin.Plugin
//	mu        sync.RWMutex
//	name      string
//	version   string
//}
//
//func NewGoRuntime() *GoRuntime {
//	return &GoRuntime{
//		functions: make(map[string]*plugin.Plugin),
//		name:      "go",
//		version:   "1.0.0",
//	}
//}
//
//func (g *GoRuntime) GetName() string {
//	return g.name
//}
//func (g *GoRuntime) GetVersion() string {
//	return g.version
//}
//
//func (g *GoRuntime) Execute(ctx context.Context, metadata FunctionMetadata, payload []byte) ([]byte, error) {
//	p, err := g.LoadPlugin(metadata.Name)
//	if err != nil {
//		return nil, fmt.Errorf("failed to load plugin: %w", err)
//	}
//	handlerName := metadata.Handler
//	f, err := p.Lookup(handlerName)
//	if err != nil {
//		return nil, fmt.Errorf("failed to lookup handler: %w", err)
//	}
//	handler, ok := f.(func(context.Context, []byte) ([]byte, error))
//	if !ok {
//		return nil, fmt.Errorf("%s is not of type func(context.Context, []byte) ([]byte, error)", handlerName)
//	}
//	return handler(ctx, payload)
//
//}
//
//func (g *GoRuntime) Init() error {
//	return nil
//}
//
//func (g *GoRuntime) Shutdown() error {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	g.functions = make(map[string]*plugin.Plugin)
//	return nil
//}
//
//func (g *GoRuntime) LoadPlugin(name string) (*plugin.Plugin, error) {
//	g.mu.RLock()
//	p, exists := g.functions[name]
//	g.mu.RUnlock()
//	if exists {
//		return p, nil
//	}
//
//	g.mu.Lock()
//	defer g.mu.Unlock()
//
//	if p, exists = g.functions[name]; exists {
//		return p, nil
//	}
//
//	// Load the plugin
//	goPluginPath := fmt.Sprintf("./functions/go/%s.so", name)
//	if !common.FileExists(goPluginPath) {
//		return nil, fmt.Errorf("plugin %s does not exist", goPluginPath)
//	}
//	p, err := plugin.Open(goPluginPath)
//	if err != nil {
//		return nil, err
//	}
//
//	g.functions[name] = p
//	return p, nil
//}
//
//func (g *GoRuntime) Validate(metadata FunctionMetadata) error {
//	if metadata.Runtime != g.name {
//		return fmt.Errorf("invalid runtime %s, expected %s", metadata.Runtime, g.name)
//	}
//	if metadata.CodePath == "" {
//		return fmt.Errorf("code path is required")
//	}
//	if metadata.Handler == "" {
//		return fmt.Errorf("handler is required")
//	}
//	if !common.FileExists(metadata.CodePath) {
//		return fmt.Errorf("code path %s does not exist", metadata.CodePath)
//	}
//	return nil
//}
//
//func (g *GoRuntime) Prepare(metadata FunctionMetadata) error {
//	// Check if the path is a plugin or a go file or a directory
//	if common.IsPlugin(metadata.CodePath) {
//		// Copy the plugin to the functions directory
//		goPluginPath := fmt.Sprintf("./functions/go/%s.so", metadata.Name)
//		if err := common.CopyFile(metadata.CodePath, goPluginPath); err != nil {
//			return fmt.Errorf("failed to copy plugin: %w", err)
//		}
//		return nil
//	}
//	if common.IsGoFile(metadata.CodePath) || common.IsDirectory(metadata.CodePath) {
//		// Compile the go file
//		goPluginPath := fmt.Sprintf("./functions/go/%s.so", metadata.Name)
//		if err := common.CompileGoFile(metadata.CodePath, goPluginPath); err != nil {
//			return fmt.Errorf("failed to compile go file: %w", err)
//		}
//		return nil
//	}
//	return fmt.Errorf("invalid code path %s", metadata.CodePath)
//}
