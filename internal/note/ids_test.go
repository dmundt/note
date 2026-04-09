package note

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunIDsPrintsSortedUniqueIDs(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410101010-a.md"), "---\nid: 20260410101010\ntitle: A\ntype: permanent\ncreated: 2026-04-10T10:10:10Z\ntags: []\n---\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "20260410111111-b.md"), "---\nid: 20260410111111\ntitle: B\ntype: fleeting\ncreated: 2026-04-10T11:11:11Z\ntags: []\n---\n")

	app := workspace.app(t)
	out, err := captureStdout(t, func() error {
		return app.runIDs(nil)
	})
	if err != nil {
		t.Fatalf("runIDs error: %v", err)
	}
	lines := splitNonEmptyLines(out)
	if len(lines) != 2 {
		t.Fatalf("expected 2 IDs, got %d: %q", len(lines), out)
	}
	if lines[0] != "20260410111111" || lines[1] != "20260410101010" {
		t.Fatalf("unexpected ID order: %v", lines)
	}
}

func TestRunIDsRejectsArgs(t *testing.T) {
	workspace := newTestWorkspace(t)
	app := workspace.app(t)
	if err := app.runIDs([]string{"extra"}); err == nil {
		t.Fatal("expected usage error when args are provided")
	}
}

func TestCompletionContainsDynamicIDSource(t *testing.T) {
	if !strings.Contains(bashCompletion, "note __ids") {
		t.Fatal("bash completion should query note __ids")
	}
	if !strings.Contains(zshCompletion, "note __ids") {
		t.Fatal("zsh completion should query note __ids")
	}
	if !strings.Contains(powershellCompletion, "note __ids") {
		t.Fatal("powershell completion should query note __ids")
	}
	if !strings.Contains(fishCompletion, "__note_ids") {
		t.Fatal("fish completion should define/use __note_ids")
	}
}

func splitNonEmptyLines(s string) []string {
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}
