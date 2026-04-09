package note

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpOutputGoldenFile(t *testing.T) {
	t.Parallel()
	workspace := newTestWorkspace(t)

	output, err := captureStdout(t, func() error {
		var runErr error
		workspace.withWorkingDirectory(t, func() {
			runErr = Run([]string{"help"})
		})
		return runErr
	})
	if err != nil {
		t.Fatalf("Run(help) returned error: %v", err)
	}
	assertGoldenFile(t, "testdata/help_output.golden.txt", output)
}

func TestSearchOutputGoldenFiles(t *testing.T) {
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
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081529-capture-idea.md"), `---
id: 202604081529
title: Capture Idea
type: fleeting
created: 2026-04-08T15:29:00Z
tags: [inbox]
source:
---

Remember this.
`)

	app := workspace.app(t)
	results, err := app.searchNotes("", "", "", "")
	if err != nil {
		t.Fatalf("searchNotes returned error: %v", err)
	}
	output, err := formatSearchResults(workspace.root, results, "pipe")
	if err != nil {
		t.Fatalf("formatSearchResults returned error: %v", err)
	}
	assertGoldenFile(t, "testdata/search_output.golden.txt", output)
}

func TestStatsOutputGoldenFile(t *testing.T) {
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
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081529-capture-idea.md"), `---
id: 202604081529
title: Capture Idea
type: fleeting
created: 2026-04-08T15:29:00Z
tags: [inbox]
source:
---

Remember this.
`)

	app := workspace.app(t)
	stats, err := app.collectStats()
	if err != nil {
		t.Fatalf("collectStats returned error: %v", err)
	}
	assertGoldenFile(t, "testdata/stats_output.golden.txt", formatStats(stats))
}

func TestCreateNoteGoldenFiles(t *testing.T) {
	t.Parallel()

	workspace := newTestWorkspace(t)
	workspace.cfg.IDFormat = "NOTEID"
	app := workspace.app(t)

	tests := []struct {
		name       string
		args       []string
		goldenPath string
	}{
		{
			name:       "fleeting",
			args:       []string{"fleeting", "Capture Idea"},
			goldenPath: "testdata/create_fleeting.golden.md",
		},
		{
			name:       "permanent",
			args:       []string{"permanent", "Linking Notes"},
			goldenPath: "testdata/create_permanent.golden.md",
		},
		{
			name:       "project",
			args:       []string{"project", "project_slug", "Project Notes"},
			goldenPath: "testdata/create_project.golden.md",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			path, err := app.createNote(test.args)
			if err != nil {
				t.Fatalf("createNote returned error: %v", err)
			}
			content := readFileForTest(t, path)
			assertGoldenFile(t, test.goldenPath, normalizeDynamicNoteContent(content))
		})
	}
}

func TestNoteRewriteGoldenFiles(t *testing.T) {
	t.Parallel()

	t.Run("title rewrite", func(t *testing.T) {
		workspace := newTestWorkspace(t)
		writeTestNote(t, workspace.root, "notes/fleeting/202604081603-smoke-test-capture.md", `---
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
			t.Fatalf("unexpected rewritten path: %s", newPath)
		}
		content := readFileForTest(t, newPath)
		assertGoldenFile(t, "testdata/rewrite_title.golden.md", normalizeDynamicNoteContent(content))
	})

	t.Run("tag rewrite", func(t *testing.T) {
		workspace := newTestWorkspace(t)
		writeTestNote(t, workspace.root, "notes/fleeting/202604081603-smoke-test-capture.md", `---
id: 202604081603
title: Smoke Test Capture
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
		if _, err := app.metaRemoveTag("202604081603", "inbox"); err != nil {
			t.Fatalf("metaRemoveTag returned error: %v", err)
		}

		content := readFileForTest(t, filepath.Join(workspace.root, "notes", "fleeting", "202604081603-smoke-test-capture.md"))
		assertGoldenFile(t, "testdata/rewrite_tags.golden.md", normalizeDynamicNoteContent(content))
	})

	t.Run("promote project note", func(t *testing.T) {
		workspace := newTestWorkspace(t)
		writeTestNote(t, workspace.root, "notes/project/202604081612-sf1-assembler4.md", `---
id: 202604081612
title: sf1 assembler4
type: project
project: sf1
created: 2026-04-08T16:12:00Z
tags: [assembler, sf1]
---

Promote this into the permanent graph.
`)

		app := workspace.app(t)
		newPath, err := app.promoteNote("202604081612", "")
		if err != nil {
			t.Fatalf("promoteNote returned error: %v", err)
		}
		if !strings.HasSuffix(filepath.ToSlash(newPath), "/notes/permanent/202604081612-sf1-assembler4.md") {
			t.Fatalf("unexpected promoted path: %s", newPath)
		}

		content := readFileForTest(t, newPath)
		assertGoldenFile(t, "testdata/promote_project_to_permanent.golden.md", normalizeDynamicNoteContent(content))
	})

	t.Run("promote with title", func(t *testing.T) {
		workspace := newTestWorkspace(t)
		writeTestNote(t, workspace.root, "notes/project/202604081612-sf1-assembler4.md", `---
id: 202604081612
title: sf1 assembler4
type: project
project: sf1
created: 2026-04-08T16:12:00Z
tags: [assembler, sf1]
---

Promote this into the permanent graph.
`)

		app := workspace.app(t)
		newPath, err := app.promoteNote("202604081612", "Assembler Insight")
		if err != nil {
			t.Fatalf("promoteNote with title returned error: %v", err)
		}
		if !strings.HasSuffix(filepath.ToSlash(newPath), "/notes/permanent/202604081612-assembler-insight.md") {
			t.Fatalf("unexpected promoted path: %s", newPath)
		}

		content := readFileForTest(t, newPath)
		assertGoldenFile(t, "testdata/promote_with_title.golden.md", normalizeDynamicNoteContent(content))
	})

	t.Run("promote already permanent", func(t *testing.T) {
		workspace := newTestWorkspace(t)
		writeTestNote(t, workspace.root, "notes/permanent/202604081700-existing-permanent-note.md", `---
id: 202604081700
title: Existing Permanent Note
type: permanent
created: 2026-04-08T17:00:00Z
tags: [zettel]
---

This note is already permanent.
`)

		app := workspace.app(t)
		newPath, err := app.promoteNote("202604081700", "")
		if err != nil {
			t.Fatalf("promoteNote on permanent note returned error: %v", err)
		}
		if !strings.HasSuffix(filepath.ToSlash(newPath), "/notes/permanent/202604081700-existing-permanent-note.md") {
			t.Fatalf("unexpected permanent path after promote: %s", newPath)
		}

		content := readFileForTest(t, newPath)
		assertGoldenFile(t, "testdata/promote_existing_permanent.golden.md", normalizeDynamicNoteContent(content))
	})
}

func captureStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	runErr := run()
	_ = writer.Close()
	os.Stdout = originalStdout

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	_ = reader.Close()
	return buffer.String(), runErr
}
