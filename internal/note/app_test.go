package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateNoteUsesTemplateAndWritesFile(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	app := workspace.app(t)

	path, err := app.createNote([]string{"fleeting", "Capture Idea"})
	if err != nil {
		t.Fatalf("createNote returned error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created note: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "title: Capture Idea") {
		t.Fatalf("created note missing title: %q", text)
	}
	if !strings.Contains(text, "type: fleeting") {
		t.Fatalf("created note missing type: %q", text)
	}
	if filepath.Dir(path) != filepath.Join(workspace.root, "notes", "fleeting") {
		t.Fatalf("note created in wrong directory: %s", path)
	}
}

func TestSearchNotesFiltersByTypeProjectTagAndQuery(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081530-linking-notes.md"), `---
id: 202604081530
title: Linking Notes
type: permanent
created: 2026-04-08T15:30:00Z
tags: [zettel, links]
---

Connect ideas with explicit IDs.
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "project", "202604081531-release-plan.md"), `---
id: 202604081531
title: Release Plan
type: project
project: launch
created: 2026-04-08T15:31:00Z
tags: [planning]
---

Coordinate launch tasks.
`)

	app := workspace.app(t)

	results, err := app.searchNotes("permanent", "", "links", "explicit")
	if err != nil {
		t.Fatalf("searchNotes returned error: %v", err)
	}
	if len(results) != 1 || results[0].Metadata["id"] != "202604081530" {
		t.Fatalf("unexpected permanent search results: %#v", results)
	}

	results, err = app.searchNotes("project", "launch", "planning", "launch")
	if err != nil {
		t.Fatalf("searchNotes project filter returned error: %v", err)
	}
	if len(results) != 1 || results[0].Metadata["id"] != "202604081531" {
		t.Fatalf("unexpected project search results: %#v", results)
	}
}

func TestSearchNotesSeesNewNoteAfterCreate(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	app := workspace.app(t)

	initial, err := app.searchNotes("", "", "", "")
	if err != nil {
		t.Fatalf("initial searchNotes returned error: %v", err)
	}
	if len(initial) != 0 {
		t.Fatalf("expected empty note set at start, got %d", len(initial))
	}

	if _, err := app.createNote([]string{"fleeting", "Index Invalidation"}); err != nil {
		t.Fatalf("createNote returned error: %v", err)
	}

	results, err := app.searchNotes("", "", "", "invalidation")
	if err != nil {
		t.Fatalf("searchNotes after create returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result after create, got %d", len(results))
	}
}
