package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type ReadFileTool struct{}

func NewReadFileTool() *ReadFileTool { return &ReadFileTool{} }

func (r *ReadFileTool) Name() string { return "read_file" }

func (r *ReadFileTool) Description() string {
	return "Read the contents of a file from the filesystem. Returns the file content as text."
}

func (r *ReadFileTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to read.",
			},
		},
		"required": []string{"path"},
	}
}

func (r *ReadFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(params.Path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
