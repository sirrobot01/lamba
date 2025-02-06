package lang

import (
	"fmt"
	"github.com/sirrobot01/lamba/internal/event"
	"github.com/sirrobot01/lamba/internal/function"
)

func (r *Runtime) GetNodeJSCmd(event *event.Event, fn *function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()

	nodeCmd := `
	const { %s } = require('./%s.js');

	// Capture console.log output
	const logs = [];
	const originalConsoleLog = console.log;
	console.log = (...args) => {
		logs.push(args.map(arg => String(arg)).join(' '));
	};
	
	const eventData = JSON.parse('%s');
	const fnData = JSON.parse('%s');
	async function main() {
		const result = await %s(eventData, fnData);
		
		// Restore console.log
		console.log = originalConsoleLog;
		
		// Output final result with captured logs
		console.log(JSON.stringify({
			result,
			debug: logs
		}));
	}
	
	main().catch(error => {
		console.error(error);
		process.exit(1);
	});
`

	nodeCmd = fmt.Sprintf(nodeCmd, fn.Handler, fn.Name, eventJson, fnJSON, fn.Handler)
	return []string{"nodejs", "-e", nodeCmd}
}
