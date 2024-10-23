package common

import "encoding/json"

func ParsePayload(payload string) any {
	var jsonPayload any
	if err := json.Unmarshal([]byte(payload), &jsonPayload); err != nil {
		// If not valid JSON, return as string
		return payload
	}
	return jsonPayload
}
