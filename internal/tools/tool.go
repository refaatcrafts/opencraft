package tools

import (
	"context"
	"encoding/json"
)

// Tool is the interface all tools must implement.
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}
