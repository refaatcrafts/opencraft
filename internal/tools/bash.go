package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type BashTool struct{}

func NewBashTool() *BashTool { return &BashTool{} }

func (b *BashTool) Name() string { return "bash" }

func (b *BashTool) Description() string {
	return "Execute a bash shell command and return its output. Use for running commands, scripts, or any shell operations."
}

func (b *BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The bash command to execute.",
			},
		},
		"required": []string{"command"},
	}
}

func (b *BashTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if params.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", params.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String()
	if stderr.Len() > 0 {
		if out != "" {
			out += "\n"
		}
		out += stderr.String()
	}

	if err != nil {
		if out != "" {
			return out, fmt.Errorf("%w", err)
		}
		return "", err
	}
	return out, nil
}
