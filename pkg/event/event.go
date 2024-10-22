package event

import (
	"encoding/json"
	"strings"
	"time"
)

type InvokeEvent struct {
	Name      string
	Payload   []byte
	InvokedAt time.Time
}

func (i *InvokeEvent) ToJSON() string {
	j, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return strings.Replace(string(j), "'", "\\'", -1)
}
