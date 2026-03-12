package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	baseSystemInstruction = "You are opencraft, a helpful AI coding assistant. You have access to tools to read files, write files, list directories, and run bash commands. Be concise and helpful."
	guidelinesStartTag    = "<INSTRUCTIONS>"
	guidelinesEndTag      = "</INSTRUCTIONS>"
	maxGuidelinesChars    = 12000
)

func (a *Agent) systemInstruction() (instruction string, guidelinesPath string) {
	path, guidelines, err := loadNearestAGENTSInstructions(".")
	if err != nil || guidelines == "" {
		return baseSystemInstruction, ""
	}

	guidelines = truncateGuidelines(guidelines, maxGuidelinesChars)
	return fmt.Sprintf(
		"%s\n\nRepository instructions from %s:\n%s",
		baseSystemInstruction,
		path,
		guidelines,
	), path
}

func loadNearestAGENTSInstructions(startDir string) (string, string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", "", err
	}

	for {
		path := filepath.Join(abs, "AGENTS.md")
		data, readErr := os.ReadFile(path)
		switch {
		case readErr == nil:
			return path, extractGuidelines(string(data)), nil
		case errors.Is(readErr, os.ErrNotExist):
			// Continue walking up.
		default:
			return "", "", readErr
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}

	return "", "", nil
}

func extractGuidelines(contents string) string {
	trimmed := strings.TrimSpace(contents)
	if trimmed == "" {
		return ""
	}

	start := strings.Index(trimmed, guidelinesStartTag)
	if start < 0 {
		return trimmed
	}
	start += len(guidelinesStartTag)

	endRel := strings.Index(trimmed[start:], guidelinesEndTag)
	if endRel < 0 {
		return trimmed
	}

	return strings.TrimSpace(trimmed[start : start+endRel])
}

func truncateGuidelines(guidelines string, limit int) string {
	if limit <= 0 || len(guidelines) <= limit {
		return guidelines
	}
	const suffix = "\n\n[AGENTS.md was truncated to fit prompt limits.]"
	trimmed := strings.TrimRight(guidelines[:limit], "\n\r\t ")
	return trimmed + suffix
}
