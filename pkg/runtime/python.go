package runtime

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

type PythonRuntime struct {
	ContainerdRuntime
}

func NewPythonRuntime() *PythonRuntime {
	runtime, err := NewContainerdRuntime(
		"python",
		"python:3.9-alpine",
		"3.9",
	)
	if err != nil {
		return nil
	}

	return &PythonRuntime{
		ContainerdRuntime: *runtime,
	}
}

func (r *PythonRuntime) GetCmd(event *event.Event, fn *function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()
	pythonCmd := `
import json
import sys
from io import StringIO
from %s import %s

# Capture prints
captured_output = StringIO()
sys.stdout = captured_output

# Run function
result = %s(json.loads('%s'), json.loads('%s'))

# Get printed output
prints = captured_output.getvalue()

# Restore stdout
sys.stdout = sys.__stdout__
print(json.dumps({
    "result": result,
    "debug": prints.split("\n")
}))
`
	pythonCmd = fmt.Sprintf(pythonCmd, fn.Name, fn.Handler, fn.Handler, eventJson, fnJSON)
	return []string{"python", "-c", pythonCmd}
}
