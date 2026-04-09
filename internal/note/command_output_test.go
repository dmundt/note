package note

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSearchOutputsTable(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081930-linking-notes.md"), `---
id: 202604081930
title: Linking Notes
type: permanent
created: 2026-04-08T19:30:00Z
tags: [links]
---

Connect ideas with explicit links.
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "project", "202604081931-release-links.md"), `---
id: 202604081931
title: Release Links
type: project
project: launch
created: 2026-04-08T19:31:00Z
tags: [planning]
---

Coordinate release links.
`)

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runSearch([]string{"--format", "table", "links"})
	})
	if err != nil {
		t.Fatalf("runSearch returned error: %v", err)
	}
	if !strings.Contains(output, "id") || !strings.Contains(output, "title") || !strings.Contains(output, "path") {
		t.Fatalf("table output missing header columns: %q", output)
	}
	if !strings.Contains(output, "notes/permanent/202604081930-linking-notes.md") {
		t.Fatalf("table output missing expected result: %q", output)
	}
}

func TestRunSearchOutputsJSONForNoMatches(t *testing.T) {
	workspace := newTestWorkspace(t)
	app := workspace.app(t)

	output, err := captureStdout(t, func() error {
		return app.runSearch([]string{"--format", "json", "definitely-no-match"})
	})
	if err != nil {
		t.Fatalf("runSearch returned error: %v", err)
	}
	if output != "[]\n" {
		t.Fatalf("unexpected JSON output for empty result: %q", output)
	}
}

func TestRunSearchUsesConfigDefaultOutputFormat(t *testing.T) {
	workspace := newTestWorkspace(t)
	workspace.cfg.SearchOutputDefault = "json"
	app := workspace.app(t)

	output, err := captureStdout(t, func() error {
		return app.runSearch([]string{"definitely-no-match"})
	})
	if err != nil {
		t.Fatalf("runSearch returned error: %v", err)
	}
	if output != "[]\n" {
		t.Fatalf("expected JSON output from default format, got: %q", output)
	}
}

func TestRunSearchStrictFailsOnMalformedNote(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081939-bad.md"), `---
id: 202604081939
title malformed
type: permanent
created: 2026-04-08T19:39:00Z
tags: []
---

broken
`)

	app := workspace.app(t)
	err := app.runSearch([]string{"--strict", "anything"})
	if err == nil {
		t.Fatal("expected strict search to fail on malformed note")
	}
	if !strings.Contains(err.Error(), "strict mode") || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("unexpected strict search error: %v", err)
	}
}

func TestRunBacklinksOutputOrder(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081940-target.md"), `---
id: 202604081940
title: Target
type: permanent
created: 2026-04-08T19:40:00Z
tags: []
---

Target note.
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "project", "202604081942-link-newer.md"), `---
id: 202604081942
title: Link Newer
type: project
project: p1
created: 2026-04-08T19:42:00Z
tags: []
---

See [target](../permanent/202604081940-target.md)
`)
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081941-link-older.md"), `---
id: 202604081941
title: Link Older
type: fleeting
created: 2026-04-08T19:41:00Z
tags: []
source:
---

See [[202604081940]]
`)

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runBacklinks([]string{"202604081940"})
	})
	if err != nil {
		t.Fatalf("runBacklinks returned error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected backlink output lines: %v", lines)
	}
	if !strings.HasPrefix(lines[0], "202604081942") || !strings.HasPrefix(lines[1], "202604081941") {
		t.Fatalf("unexpected backlink order/output: %v", lines)
	}
}

func TestRunDoctorOKOutput(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081950-valid.md"), `---
id: 202604081950
title: Valid
type: permanent
created: 2026-04-08T19:50:00Z
tags: []
---

Valid note.
`)

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runDoctor(nil)
	})
	if err != nil {
		t.Fatalf("runDoctor returned error: %v", err)
	}
	if !strings.Contains(output, "doctor: ok (1 notes checked)") {
		t.Fatalf("unexpected doctor ok output: %q", output)
	}
}

func TestRunDoctorReportsIssues(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604082000-broken.md"), `---
id: 202604082000
title: Broken
type: permanent
created: 2026-04-08T20:00:00Z
tags: []
---

Broken [link](./missing.md)
`)

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runDoctor(nil)
	})
	if err == nil {
		t.Fatal("expected runDoctor to return error when issues are present")
	}
	if !strings.Contains(err.Error(), "doctor found issues") {
		t.Fatalf("unexpected runDoctor error: %v", err)
	}
	if !strings.Contains(output, "doctor: found 1 issue(s)") || !strings.Contains(output, "broken markdown link") {
		t.Fatalf("unexpected doctor issue output: %q", output)
	}
}
