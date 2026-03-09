package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type ListDirTool struct{}

func NewListDirTool() *ListDirTool { return &ListDirTool{} }

func (l *ListDirTool) Name() string { return "list_dir" }

func (l *ListDirTool) Description() string {
	return "List the contents of a directory. Returns file names, types, and sizes."
}

func (l *ListDirTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The directory path to list. Defaults to current directory.",
			},
		},
		"required": []string{},
	}
}

func (l *ListDirTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	dir := params.Path
	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	sort.Slice(entries, func(i, j int) bool {
		// directories first
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	var sb strings.Builder
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("%-40s  <dir>\n", entry.Name()+"/"))
		} else {
			sb.WriteString(fmt.Sprintf("%-40s  %d bytes\n", entry.Name(), info.Size()))
		}
	}

	return sb.String(), nil
}
