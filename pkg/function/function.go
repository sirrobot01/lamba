package function

import (
	"encoding/json"
	"strings"
)

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Runtime     string `json:"runtime"`
	Handler     string `json:"handler"`
	CodePath    string `json:"codePath"`
	PreExec     string `json:"preExec"` // Pre-execution command
	Timeout     int    `json:"timeout"`
}

func (f *Function) ToJSON() string {
	j, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return strings.Replace(string(j), "'", "\\'", -1)
}
