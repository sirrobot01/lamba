package runtime

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

type PythonRuntime struct {
	DockerRuntime
}

func NewPythonRuntime() *PythonRuntime {
	return &PythonRuntime{
		DockerRuntime: DockerRuntime{
			name:    "python",
			image:   "python:3.9-alpine",
			version: "3.9",
		},
	}
}

func (r *PythonRuntime) GetCmd(event *event.Event, fn *function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()
	pythonCmd := fmt.Sprintf(
		"import json; from %s import %s; result = %s(json.loads('%s'), json.loads('%s')); print(json.dumps(result))",
		fn.Name, fn.Handler, fn.Handler, eventJson, fnJSON,
	)
	return []string{"python", "-c", pythonCmd}
}
