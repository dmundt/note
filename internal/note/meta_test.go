package note

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMetaSetRenamesFileAndUpdatesTitle(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	originalPath := filepath.Join("notes", "fleeting", "202604081603-smoke-test-capture.md")
	writeTestNote(t, workspace.root, originalPath, `---
id: 202604081603
title: smoke test capture
type: fleeting
created: 2026-04-08T16:03:18Z
tags: []
source:
---

`)

	app := workspace.app(t)
	newPath, err := app.metaSet("202604081603", "title", "Better Capture")
	if err != nil {
		t.Fatalf("metaSet returned error: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(newPath), "/202604081603-better-capture.md") {
		t.Fatalf("unexpected renamed path: %s", newPath)
	}

	note, err := parseNoteFile(newPath)
	if err != nil {
		t.Fatalf("parse renamed note: %v", err)
	}
	if note.Metadata["title"] != "Better Capture" {
		t.Fatalf("title was not updated: %#v", note.Metadata)
	}
	if note.Metadata["updated"] == "" {
		t.Fatalf("updated timestamp was not set: %#v", note.Metadata)
	}
}

func TestMetaAddAndRemoveTag(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081603-smoke-test-capture.md"), `---
id: 202604081603
title: smoke test capture
type: fleeting
created: 2026-04-08T16:03:18Z
tags: [inbox]
source:
---

`)

	app := workspace.app(t)
	if _, err := app.metaAddTag("202604081603", "zettel"); err != nil {
		t.Fatalf("metaAddTag returned error: %v", err)
	}
	if _, err := app.metaAddTag("202604081603", "ZETTEL"); err != nil {
		t.Fatalf("metaAddTag duplicate returned error: %v", err)
	}
	if _, err := app.metaRemoveTag("202604081603", "inbox"); err != nil {
		t.Fatalf("metaRemoveTag returned error: %v", err)
	}

	note, err := parseNoteFile(filepath.Join(workspace.root, "notes", "fleeting", "202604081603-smoke-test-capture.md"))
	if err != nil {
		t.Fatalf("parse note after tag edits: %v", err)
	}
	if len(note.Tags) != 1 || note.Tags[0] != "zettel" {
		t.Fatalf("unexpected tags after edits: %#v", note.Tags)
	}
	if note.Metadata["updated"] == "" {
		t.Fatalf("updated timestamp was not set: %#v", note.Metadata)
	}
}
