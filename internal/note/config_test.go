package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppDefaultsSearchOutput(t *testing.T) {
	workspace := t.TempDir()
	writeTestNote(t, workspace, "note.json", `{
  "version": 1,
  "idFormat": "20060102150405",
  "filenamePattern": "{id}-{slug}.md",
  "linkStyle": "wiki-id",
  "noteTypes": {
    "fleeting": {"dir": "notes/fleeting", "template": "templates/fleeting.md", "frontmatter": ["id", "title", "type", "created", "tags", "source"]}
  }
}`)

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	defer func() { _ = os.Chdir(original) }()

	app, err := loadApp()
	if err != nil {
		t.Fatalf("loadApp returned error: %v", err)
	}
	if app.cfg.SearchOutputDefault != "pipe" {
		t.Fatalf("expected default search output format pipe, got %q", app.cfg.SearchOutputDefault)
	}
}

func TestLoadAppRejectsInvalidSearchOutput(t *testing.T) {
	workspace := t.TempDir()
	writeTestNote(t, workspace, "note.json", `{
  "version": 1,
  "idFormat": "20060102150405",
  "filenamePattern": "{id}-{slug}.md",
  "linkStyle": "wiki-id",
  "searchOutputDefault": "xml",
  "noteTypes": {
    "fleeting": {"dir": "notes/fleeting", "template": "templates/fleeting.md", "frontmatter": ["id", "title", "type", "created", "tags", "source"]}
  }
}`)

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	defer func() { _ = os.Chdir(original) }()

	_, err = loadApp()
	if err == nil {
		t.Fatal("expected loadApp to fail for invalid searchOutputDefault")
	}
	if !strings.Contains(err.Error(), "invalid searchOutputDefault") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindWorkspaceRootWalksParents(t *testing.T) {
	workspace := t.TempDir()
	writeTestNote(t, workspace, "note.json", `{"version":1}`)
	nested := filepath.Join(workspace, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	defer func() { _ = os.Chdir(original) }()

	root, err := findWorkspaceRoot()
	if err != nil {
		t.Fatalf("findWorkspaceRoot returned error: %v", err)
	}
	if filepath.Clean(root) != filepath.Clean(workspace) {
		t.Fatalf("unexpected workspace root: %s", root)
	}
}
