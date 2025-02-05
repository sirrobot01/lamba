package runtime

import (
	"github.com/sirrobot01/lamba/pkg/runtime/engines"
	"github.com/sirrobot01/lamba/pkg/runtime/languages"
)

func GetRuntimes(engine engines.Type) map[string]languages.Runtime {
	return map[string]languages.Runtime{
		"golang": languages.NewRuntime(engine, "golang", "golang:1.22-alpine", "1.22"),
		"node":   languages.NewRuntime(engine, "node", "node:14-alpine", "14"),
		"python": languages.NewRuntime(engine, "python", "python:3.9-alpine", "3.9"),
	}
}
