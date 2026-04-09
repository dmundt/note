package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesWorkspace(t *testing.T) {
	dir := t.TempDir()

	if err := runInit(dir, nil); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	wantFiles := []string{"note.json"}
	for _, rel := range wantFiles {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file to exist: %s", rel)
		}
	}

	wantDirs := []string{
		filepath.Join("notes", "fleeting"),
		filepath.Join("notes", "permanent"),
		filepath.Join("notes", "project"),
	}
	for _, rel := range wantDirs {
		path := filepath.Join(dir, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected directory to exist: %s", rel)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file: %s", rel)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "templates")); !os.IsNotExist(err) {
		t.Errorf("expected templates directory to be absent by default init")
	}
}

func TestRunInitFailsIfAlreadyInitialised(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, nil); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	err := runInit(dir, nil)
	if err == nil {
		t.Fatal("expected error when running init in an already-initialised workspace")
	}
}

func TestRunInitNoteJSONIsValid(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, nil); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	// Change to the new workspace dir so loadApp can find note.json.
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	app, err := loadApp()
	if err != nil {
		t.Fatalf("loadApp on init workspace: %v", err)
	}
	if app.cfg.IDFormat != "20060102150405" {
		t.Errorf("unexpected idFormat: %q", app.cfg.IDFormat)
	}
	if len(app.cfg.NoteTypes) != 3 {
		t.Errorf("expected 3 note types, got %d", len(app.cfg.NoteTypes))
	}
}

func TestRunInitTemplatesHaveExpectedPlaceholders(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, []string{"full"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	checks := map[string]string{
		filepath.Join("templates", "fleeting.md"):  "{{id}}",
		filepath.Join("templates", "permanent.md"): "{{title}}",
		filepath.Join("templates", "project.md"):   "{{project}}",
	}
	for rel, want := range checks {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			t.Errorf("read %s: %v", rel, err)
			continue
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("%s: expected placeholder %q", rel, want)
		}
	}
}

func TestRunInitDefaultUsesEmbeddedTemplatesWhenWorkspaceTemplatesMissing(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, nil); err != nil {
		t.Fatalf("runInit default failed: %v", err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	app, err := loadApp()
	if err != nil {
		t.Fatalf("loadApp: %v", err)
	}

	path, err := app.createNote([]string{"fleeting", "fallback template note"})
	if err != nil {
		t.Fatalf("createNote with embedded fallback failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created note: %v", err)
	}
	if !strings.Contains(string(data), "type: fleeting") {
		t.Fatalf("expected note from embedded fleeting template, got:\n%s", string(data))
	}
}

func TestRunInitFullCreatesTemplatesDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, []string{"full"}); err != nil {
		t.Fatalf("runInit full failed: %v", err)
	}

	for _, rel := range []string{
		filepath.Join("templates", "fleeting.md"),
		filepath.Join("templates", "permanent.md"),
		filepath.Join("templates", "project.md"),
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("expected template file to exist in full init: %s", rel)
		}
	}
}

func TestRunInitRejectsUnknownArgument(t *testing.T) {
	dir := t.TempDir()
	if err := runInit(dir, []string{"weird"}); err == nil {
		t.Fatal("expected usage error for unknown init arg")
	}
}
