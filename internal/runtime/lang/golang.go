package lang

import (
	"fmt"
	"github.com/sirrobot01/lamba/internal/event"
	"github.com/sirrobot01/lamba/internal/function"
	"os"
	"path/filepath"
)

func (r *Runtime) GetGoCmd(event *event.Event, fn *function.Function) []string {
	eventJson := event.ToJSON()
	fnJSON := fn.ToJSON()

	mainFile := `
package main

import (
    "encoding/json"
    "fmt"
	"os"
)

func main() {
    var eventData, fnData map[string]interface{}
    if err := json.Unmarshal([]byte(` + "`" + `%s` + "`" + `), &eventData); err != nil {
		os.Exit(1)
    }
    if err := json.Unmarshal([]byte(` + "`" + `%s` + "`" + `), &fnData); err != nil {
        fmt.Printf("Error parsing function: %v\n", err)
        os.Exit(1)
    }

    result := Runner(eventData, fnData)
    output := map[string]interface{}{
        "result": result,
        "debug": []string{},
    }
    
    jsonOutput, _ := json.Marshal(output)
    fmt.Println(string(jsonOutput))
}
`
	// Write to local directory
	mainPath := filepath.Join(fn.CodePath, "main.go")
	if err := os.WriteFile(mainPath, []byte(fmt.Sprintf(mainFile, eventJson, fnJSON)), 0644); err != nil {
		return nil
	}

	// Run both files together
	return []string{"go", "run",
		filepath.Join("/app", fmt.Sprintf("%s.go", fn.Name)),
		filepath.Join("/app", "main.go"),
	}
}
