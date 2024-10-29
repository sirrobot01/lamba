package runtime

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

type GoRuntime struct {
	ContainerdRuntime
}

func NewGoRuntime() (*GoRuntime, error) {
	runtime, err := NewContainerdRuntime(
		"go",
		"golang:1.21-alpine",
		"",
	)
	if err != nil {
		return nil, err
	}

	return &GoRuntime{
		ContainerdRuntime: *runtime,
	}, nil
}

func (r *GoRuntime) GetCmd(event *event.Event, fn *function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()
	goCmd := `
package main

import (
    "encoding/json"
    "fmt"
    "os"
    %s
)

// Capture stdout
type stdoutCapture struct {
    logs []string
}

func (c *stdoutCapture) Write(p []byte) (n int, err error) {
    c.logs = append(c.logs, string(p))
    return len(p), nil
}

func main() {
    capture := &stdoutCapture{}
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    var event, fnConfig map[string]interface{}
    json.Unmarshal([]byte('%s'), &event)
    json.Unmarshal([]byte('%s'), &fnConfig)

    result := %s(event, fnConfig)

    w.Close()
    os.Stdout = oldStdout

    output := map[string]interface{}{
        "result": result,
        "debug":  capture.logs,
    }

    json.NewEncoder(os.Stdout).Encode(output)
}
`
	goCmd = fmt.Sprintf(goCmd, fn.Name, eventJson, fnJSON, fn.Handler)
	return []string{"go", "run", "-e", goCmd}
}
