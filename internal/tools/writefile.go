package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WriteFileTool struct{}

func NewWriteFileTool() *WriteFileTool { return &WriteFileTool{} }

func (w *WriteFileTool) Name() string { return "write_file" }

func (w *WriteFileTool) Description() string {
	return "Write content to a file on the filesystem. Creates the file if it does not exist, or overwrites it."
}

func (w *WriteFileTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to write.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file.",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (w *WriteFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if params.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	if err := os.MkdirAll(filepath.Dir(params.Path), 0o755); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(params.Path, []byte(params.Content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(params.Content), params.Path), nil
}
