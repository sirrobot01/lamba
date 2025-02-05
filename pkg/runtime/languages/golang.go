package languages

import (
	"github.com/sirrobot01/lamba/pkg/event"
	"github.com/sirrobot01/lamba/pkg/function"
)

func (r *Runtime) GetGoCmd(event *event.Event, fn *function.Function) []string {
	return []string{"go", "run", fn.CodePath}
}
