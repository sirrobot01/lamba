package function

import (
	"encoding/json"
	"strings"
	"time"
)

type Function struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Runtime       string `json:"runtime"`
	Handler       string `json:"handler"`
	CodePath      string `json:"codePath"`
	PreExec       string `json:"preExec"` // Pre-execution command
	Timeout       int    `json:"timeout"`
	ContainerID   string `json:"containerID"`
	ContainerName string `json:"containerName"`

	// Metadata

	LastRun time.Time `json:"lastRun"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (f *Function) ToJSON() string {
	if f == nil {
		return ""
	}

	j, err := json.Marshal(f)
	if err != nil {
		return ""
	}

	return strings.Replace(string(j), "'", "\\'", -1)
}
