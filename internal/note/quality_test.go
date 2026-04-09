package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindNoteByIDReturnsDuplicateError(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081530-first.md"), `---
id: 202604081530
title: first
type: fleeting
created: 2026-04-08T15:30:00Z
tags: []
source:
---

first
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "project", "202604081530-second.md"), `---
id: 202604081530
title: second
type: project
project: dup
created: 2026-04-08T15:31:00Z
tags: []
---

second
`)

	app := workspace.app(t)
	_, _, _, err := app.findNoteByID("202604081530")
	if err == nil {
		t.Fatal("expected duplicate id error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate note id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchNotesSupportsUnicodeQuery(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081630-siamese.md"), `---
id: 202604081630
title: Café Siamese behavior
type: permanent
created: 2026-04-08T16:30:00Z
tags: [breed]
---

Siamese cat in a café context.
`)

	app := workspace.app(t)
	results, err := app.searchNotes("", "", "", "café")
	if err != nil {
		t.Fatalf("searchNotes returned error: %v", err)
	}
	if len(results) != 1 || results[0].Metadata["id"] != "202604081630" {
		t.Fatalf("unexpected unicode search results: %#v", results)
	}
}

func TestDoctorFindsBrokenLinksAndUnknownDir(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081700-root.md"), `---
id: 202604081700
title: Root
type: permanent
created: 2026-04-08T17:00:00Z
tags: [links]
---

Broken link to [missing](./does-not-exist.md)
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "archive", "202604081701-archived.md"), `---
id: 202604081701
title: archived
type: permanent
created: 2026-04-08T17:01:00Z
tags: []
---

archived
`)

	app := workspace.app(t)
	issues, checked, err := app.collectDoctorIssues()
	if err != nil {
		t.Fatalf("collectDoctorIssues returned error: %v", err)
	}
	if checked == 0 {
		t.Fatal("expected checked notes count > 0")
	}

	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, "broken markdown link") {
		t.Fatalf("expected broken markdown link issue, got: %s", joined)
	}
	if !strings.Contains(joined, "unknown notes directory") {
		t.Fatalf("expected unknown directory issue, got: %s", joined)
	}
}

func TestFormatSearchResultsJSON(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081702-example.md"), `---
id: 202604081702
title: Example
type: fleeting
created: 2026-04-08T17:02:00Z
tags: []
source:
---

body
`)

	app := workspace.app(t)
	results, err := app.searchNotes("", "", "", "")
	if err != nil {
		t.Fatalf("searchNotes returned error: %v", err)
	}
	output, err := formatSearchResults(workspace.root, results, "json")
	if err != nil {
		t.Fatalf("formatSearchResults returned error: %v", err)
	}
	if !strings.Contains(output, `"id": "202604081702"`) {
		t.Fatalf("expected JSON output with id, got: %s", output)
	}
}

func TestBacklinksFindsMarkdownAndWikiLinks(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081800-target-note.md"), `---
id: 202604081800
title: Target Note
type: permanent
created: 2026-04-08T18:00:00Z
tags: []
---

target
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "project", "202604081801-project-link.md"), `---
id: 202604081801
title: Project Link
type: project
project: p1
created: 2026-04-08T18:01:00Z
tags: []
---

See [target](../permanent/202604081800-target-note.md).
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081802-wiki-link.md"), `---
id: 202604081802
title: Wiki Link
type: fleeting
created: 2026-04-08T18:02:00Z
tags: []
source:
---

Reference [[202604081800]].
`)

	app := workspace.app(t)
	results, err := app.backlinks("202604081800")
	if err != nil {
		t.Fatalf("backlinks returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("unexpected backlink count: %d", len(results))
	}
}

func TestFormatSearchResultsEmptyJSON(t *testing.T) {
	t.Parallel()

	output, err := formatSearchResults(".", nil, "json")
	if err != nil {
		t.Fatalf("formatSearchResults returned error: %v", err)
	}
	if output != "[]\n" {
		t.Fatalf("unexpected JSON output: %q", output)
	}
}

func TestFormatSearchResultsTableHasHeader(t *testing.T) {
	t.Parallel()

	results := []Note{
		{
			Path: "notes/permanent/202604081900-a.md",
			Metadata: map[string]string{
				"id":    "202604081900",
				"type":  "permanent",
				"title": "A",
			},
		},
	}

	output, err := formatSearchResults(".", results, "table")
	if err != nil {
		t.Fatalf("formatSearchResults returned error: %v", err)
	}
	if !strings.Contains(output, "id") || !strings.Contains(output, "type") || !strings.Contains(output, "title") || !strings.Contains(output, "path") {
		t.Fatalf("table output missing header columns: %q", output)
	}
}

func TestDoctorAllowsMissingOptionalFields(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081910-optional-source.md"), `---
id: 202604081910
title: Optional Source
type: fleeting
created: 2026-04-08T19:10:00Z
tags: []
---

body
`)

	app := workspace.app(t)
	issues, _, err := app.collectDoctorIssues()
	if err != nil {
		t.Fatalf("collectDoctorIssues returned error: %v", err)
	}
	for _, issue := range issues {
		if strings.Contains(issue, "missing required field \"source\"") || strings.Contains(issue, "missing required field \"updated\"") {
			t.Fatalf("doctor reported optional field as required: %s", issue)
		}
	}
}

func TestDoctorReportsMalformedNotes(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081911-bad.md"), `---
id: 202604081911
title malformed
type: permanent
created: 2026-04-08T19:11:00Z
tags: []
---

broken
`)

	app := workspace.app(t)
	issues, _, err := app.collectDoctorIssues()
	if err != nil {
		t.Fatalf("collectDoctorIssues returned error: %v", err)
	}
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, "malformed note:") {
		t.Fatalf("expected malformed note issue, got: %s", joined)
	}
}

func TestMatchLinkToTargetVariants(t *testing.T) {
	t.Parallel()

	targetID := "202604081800"
	targetFile := "202604081800-target-note.md"
	targetPath := filepath.Clean(filepath.Join("notes", "permanent", targetFile))
	candidate := Note{Path: filepath.Join("notes", "project", "202604081801-project-link.md")}

	candidate.Body = "See [[202604081800]]"
	if !hasBacklinkTo(candidate, targetID, targetFile, targetPath) {
		t.Fatal("expected bare wiki link to match")
	}
	candidate.Body = "See [[202604081800|Target Note]]"
	if !hasBacklinkTo(candidate, targetID, targetFile, targetPath) {
		t.Fatal("expected piped wiki link to match")
	}
	candidate.Body = "See [target](./202604081800-target-note.md)"
	if !hasBacklinkTo(candidate, targetID, targetFile, targetPath) {
		t.Fatal("expected relative markdown link to match")
	}
	candidate.Body = "See [target](../permanent/202604081800-target-note.md)"
	if !hasBacklinkTo(candidate, targetID, targetFile, targetPath) {
		t.Fatal("expected parent-relative markdown link to match")
	}
	candidate.Body = "No link here"
	if hasBacklinkTo(candidate, targetID, targetFile, targetPath) {
		t.Fatal("did not expect non-link content to match")
	}
}

func TestWriteNoteFileLeavesNoTempFileOnOverwrite(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	path := filepath.Join(workspace.root, "notes", "permanent", "202604081920-temp-check.md")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081920-temp-check.md"), `---
id: 202604081920
title: Temp Check
type: permanent
created: 2026-04-08T19:20:00Z
tags: []
---

body
`)

	note, err := parseNoteFile(path)
	if err != nil {
		t.Fatalf("parseNoteFile returned error: %v", err)
	}
	note.Metadata["title"] = "Temp Check Updated"
	typeCfg := workspace.cfg.NoteTypes["permanent"]
	if err := writeNoteFile(path, path, note, typeCfg); err != nil {
		t.Fatalf("writeNoteFile returned error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".note-tmp-") || strings.HasSuffix(name, ".bak") {
			t.Fatalf("found unexpected temp/backup artifact: %s", name)
		}
	}
}
