package examples

import (
	"context"
	"fmt"
)

func Hello(ctx context.Context, payload []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("Hello, %s", string(payload))), nil
}
