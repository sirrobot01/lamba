package languages

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

func (r *Runtime) GetPythonCmd(event *event.Event, fn *function.Function) []string {
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
	return []string{"/usr/local/bin/python3", "-c", pythonCmd}
}
