package runtime

import (
	"github.com/sirrobot01/lamba/internal/engines"
	"github.com/sirrobot01/lamba/internal/runtime/lang"
)

func GetRuntimes(engine engines.Type) map[string]lang.Runtime {
	return map[string]lang.Runtime{
		"golang": lang.NewRuntime(engine, "golang", "golang:1.22-alpine", "1.22"),
		"nodejs": lang.NewRuntime(engine, "nodejs", "node:14-alpine", "14"),
		"python": lang.NewRuntime(engine, "python", "python:3.9-alpine", "3.9"),
	}
}
