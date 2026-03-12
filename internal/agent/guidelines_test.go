package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractGuidelinesBlock(t *testing.T) {
	input := `
# AGENTS.md instructions for demo

<INSTRUCTIONS>
## Scope
- keep things focused
</INSTRUCTIONS>
`

	got := extractGuidelines(input)
	if strings.Contains(got, "<INSTRUCTIONS>") {
		t.Fatalf("expected extracted instructions without wrapper tags, got %q", got)
	}
	if !strings.Contains(got, "## Scope") {
		t.Fatalf("missing expected section in extracted instructions: %q", got)
	}
}

func TestExtractGuidelinesFallsBackToFullFile(t *testing.T) {
	input := "# Just plain markdown\n\n- use concise responses\n"
	got := extractGuidelines(input)
	if got != strings.TrimSpace(input) {
		t.Fatalf("got %q, want full trimmed markdown content", got)
	}
}

func TestLoadNearestAGENTSInstructionsWalksParents(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	contents := "# demo\n\n<INSTRUCTIONS>\nFollow AGENTS.md\n</INSTRUCTIONS>\n"
	if err := os.WriteFile(agentsPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	path, got, err := loadNearestAGENTSInstructions(nested)
	if err != nil {
		t.Fatalf("load instructions: %v", err)
	}
	if path != agentsPath {
		t.Fatalf("path = %q, want %q", path, agentsPath)
	}
	if got != "Follow AGENTS.md" {
		t.Fatalf("instructions = %q", got)
	}
}

func TestTruncateGuidelines(t *testing.T) {
	raw := strings.Repeat("x", 64)
	got := truncateGuidelines(raw, 20)
	if len(got) <= 20 {
		t.Fatalf("expected suffix after truncation, got %q", got)
	}
	if !strings.Contains(got, "truncated") {
		t.Fatalf("expected truncation note, got %q", got)
	}
}
