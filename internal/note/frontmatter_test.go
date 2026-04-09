package note

import (
	"strings"
	"testing"
)

func TestParseFrontmatterWithBodyAndTags(t *testing.T) {
	t.Parallel()

	content := `---
id: 202604081530
title: Linking Notes by ID
type: permanent
tags: [zettel, links]
---

Body line one.
Body line two.
`

	metadata, tags, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("parseFrontmatter returned error: %v", err)
	}
	if metadata["id"] != "202604081530" {
		t.Fatalf("unexpected id: %q", metadata["id"])
	}
	if metadata["title"] != "Linking Notes by ID" {
		t.Fatalf("unexpected title: %q", metadata["title"])
	}
	if len(tags) != 2 || tags[0] != "zettel" || tags[1] != "links" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if !strings.Contains(body, "Body line one.") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestParseFrontmatterMissingClosingSeparator(t *testing.T) {
	t.Parallel()

	_, _, _, err := parseFrontmatter("---\nid: 1\ntitle: broken")
	if err == nil {
		t.Fatal("expected error for missing closing separator")
	}
}

func TestBuildFrontmatterPreservesOrderAndExtraFields(t *testing.T) {
	t.Parallel()

	note := Note{
		Metadata: map[string]string{
			"id":      "202604081530",
			"title":   "Linking Notes by ID",
			"type":    "permanent",
			"created": "2026-04-08T15:30:00Z",
			"updated": "2026-04-08T15:40:00Z",
			"source":  "book",
		},
		Tags: []string{"zettel", "links"},
		Body: "Body",
	}
	content := buildFrontmatter(note, NoteTypeConfig{Frontmatter: []string{"id", "title", "type", "created", "tags", "updated"}})

	if !strings.Contains(content, "id: 202604081530\n") {
		t.Fatalf("frontmatter missing id: %q", content)
	}
	if !strings.Contains(content, "tags: [zettel, links]\n") {
		t.Fatalf("frontmatter missing tags: %q", content)
	}
	if !strings.Contains(content, "source: book\n") {
		t.Fatalf("frontmatter missing extra metadata: %q", content)
	}
	if !strings.HasSuffix(content, "---\n") {
		t.Fatalf("frontmatter terminator incorrect: %q", content)
	}
}
