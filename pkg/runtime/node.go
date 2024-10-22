package runtime

import (
	"fmt"
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

type NodeJSRuntime struct {
	DockerRuntime
}

func NewNodeJSRuntime() *NodeJSRuntime {
	return &NodeJSRuntime{
		DockerRuntime: DockerRuntime{
			name:    "nodejs",
			image:   "node:14-alpine",
			version: "14",
		},
	}
}

func (nd *NodeJSRuntime) GetCmd(event event.InvokeEvent, fn function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()

	nodeCmd := fmt.Sprintf("const handler = require('./%s').%s; const result = handler(%s, %s); console.log(JSON.stringify(result));",
		fn.Name, fn.Handler, eventJson, fnJSON)
	return []string{"node", "-e", nodeCmd}
}
