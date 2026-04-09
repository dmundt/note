package note

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func newBenchmarkWorkspace(b *testing.B) testWorkspace {
	b.Helper()
	root := b.TempDir()
	cfg := Config{
		Version:             1,
		IDFormat:            "20060102150405",
		FilenamePattern:     "{id}-{slug}.md",
		LinkStyle:           "wiki-id",
		SearchOutputDefault: "pipe",
		NoteTypes: map[string]NoteTypeConfig{
			"fleeting": {
				Dir:         "notes/fleeting",
				Template:    "templates/fleeting.md",
				Frontmatter: []string{"id", "title", "type", "created", "tags", "source"},
			},
			"permanent": {
				Dir:         "notes/permanent",
				Template:    "templates/permanent.md",
				Frontmatter: []string{"id", "title", "type", "created", "tags", "updated"},
			},
			"project": {
				Dir:         "notes/project",
				Template:    "templates/project.md",
				Frontmatter: []string{"id", "title", "type", "project", "created", "tags", "updated"},
			},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		b.Fatalf("marshal note.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "note.json"), data, 0o644); err != nil {
		b.Fatalf("write note.json: %v", err)
	}

	writeBenchNote(b, root, filepath.Join("templates", "fleeting.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: fleeting\ncreated: {{created}}\ntags: []\nsource:\n---\n\n{{body}}\n")
	writeBenchNote(b, root, filepath.Join("templates", "permanent.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: permanent\ncreated: {{created}}\ntags: []\n---\n\n{{body}}\n")
	writeBenchNote(b, root, filepath.Join("templates", "project.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: project\nproject: {{project}}\ncreated: {{created}}\ntags: []\n---\n\n{{body}}\n")

	return testWorkspace{root: root, cfg: cfg}
}

func writeBenchNote(b *testing.B, root, relativePath, content string) {
	b.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatalf("create benchmark directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatalf("write benchmark file: %v", err)
	}
}

func BenchmarkBuildIndex1k(b *testing.B) {
	workspace := newBenchmarkWorkspace(b)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("20260408%06d", i)
		rel := filepath.Join("notes", "permanent", id+"-note.md")
		writeBenchNote(b, workspace.root, rel, fmt.Sprintf(`---
id: %s
title: Note %d
type: permanent
created: 2026-04-08T00:00:00Z
tags: [cat]
---

Body %d
`, id, i, i))
	}
	app := workspace.app(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.invalidateIndex()
		if _, err := app.getIndex(); err != nil {
			b.Fatalf("getIndex returned error: %v", err)
		}
	}
}

func BenchmarkSearchNotes1k(b *testing.B) {
	workspace := newBenchmarkWorkspace(b)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("20260408%06d", i)
		rel := filepath.Join("notes", "permanent", id+"-note.md")
		writeBenchNote(b, workspace.root, rel, fmt.Sprintf(`---
id: %s
title: Note %d
type: permanent
created: 2026-04-08T00:00:00Z
tags: [cat]
---

Siamese signal %d
`, id, i, i))
	}
	app := workspace.app(&testing.T{})
	if _, err := app.getIndex(); err != nil {
		b.Fatalf("getIndex returned error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := app.searchNotes("", "", "", "siamese")
		if err != nil {
			b.Fatalf("searchNotes returned error: %v", err)
		}
		if len(results) == 0 {
			b.Fatal("expected non-empty results")
		}
	}
}

func BenchmarkBacklinks1k(b *testing.B) {
	workspace := newBenchmarkWorkspace(b)
	targetID := "20260408111111"
	targetFile := filepath.Join("notes", "permanent", targetID+"-target.md")
	writeBenchNote(b, workspace.root, targetFile, `---
id: 20260408111111
title: Target
type: permanent
created: 2026-04-08T11:11:11Z
tags: []
---

Target note.
`)

	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("20260408%06d", i)
		rel := filepath.Join("notes", "project", id+"-link.md")
		body := "No link."
		if i%2 == 0 {
			body = "Reference [[20260408111111]] and [target](../permanent/20260408111111-target.md)"
		}
		writeBenchNote(b, workspace.root, rel, fmt.Sprintf(`---
id: %s
title: Link %d
type: project
project: p1
created: 2026-04-08T00:00:00Z
tags: []
---

%s
`, id, i, body))
	}

	app := workspace.app(&testing.T{})
	if _, err := app.getIndex(); err != nil {
		b.Fatalf("getIndex returned error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := app.backlinks(targetID)
		if err != nil {
			b.Fatalf("backlinks returned error: %v", err)
		}
		if len(results) == 0 {
			b.Fatal("expected non-empty backlinks")
		}
	}
}
