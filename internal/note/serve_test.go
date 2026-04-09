package note

import (
	"bytes"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoteIDFromLinkTarget(t *testing.T) {
	tests := []struct {
		name   string
		href   string
		wantID string
		wantOK bool
	}{
		{name: "relative markdown", href: "./20260408153042-linking-notes.md", wantID: "20260408153042", wantOK: true},
		{name: "nested markdown", href: "../permanent/20260408153042-linking-notes.md", wantID: "20260408153042", wantOK: true},
		{name: "id only", href: "20260408153042", wantID: "20260408153042", wantOK: true},
		{name: "with query", href: "./20260408153042-linking-notes.md?x=1", wantID: "20260408153042", wantOK: true},
		{name: "with hash", href: "./20260408153042-linking-notes.md#section", wantID: "20260408153042", wantOK: true},
		{name: "http external", href: "https://example.com", wantOK: false},
		{name: "anchor", href: "#local", wantOK: false},
		{name: "non-md local", href: "./guide.html", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := noteIDFromLinkTarget(tt.href)
			if gotOK != tt.wantOK {
				t.Fatalf("ok mismatch: got %v want %v", gotOK, tt.wantOK)
			}
			if gotID != tt.wantID {
				t.Fatalf("id mismatch: got %q want %q", gotID, tt.wantID)
			}
		})
	}
}

func TestEnhanceRenderedLinksRewritesInternalLinks(t *testing.T) {
	input := `<p><a href="./20260408153042-linking-notes.md">Linked</a> and <a href="https://example.com">External</a></p>`
	out := string(enhanceRenderedLinks(input, ""))

	if !strings.Contains(out, `href="/note/20260408153042"`) {
		t.Fatalf("expected rewritten href in output: %s", out)
	}
	if !strings.Contains(out, `hx-get="/note/20260408153042"`) {
		t.Fatalf("expected htmx attrs in output: %s", out)
	}
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Fatalf("expected external link untouched: %s", out)
	}
}

func TestEnhanceRenderedLinksDoesNotDoubleEscapeEntitiesWhenHighlighting(t *testing.T) {
	input := `<p>Owner absence &gt;4 hrs without stress indicators.</p>`
	out := string(enhanceRenderedLinks(input, "4"))

	if strings.Contains(out, "&amp;gt;") {
		t.Fatalf("expected no double-escaped entity, got: %s", out)
	}
	if !strings.Contains(out, "&gt;<mark>4</mark>") {
		t.Fatalf("expected highlighted number after gt entity, got: %s", out)
	}
}

func TestServeMarkdownRendersTables(t *testing.T) {
	md := newServeMarkdown()
	input := "| col | val |\n| --- | --- |\n| a | b |\n"
	var out bytes.Buffer
	if err := md.Convert([]byte(input), &out); err != nil {
		t.Fatalf("convert markdown: %v", err)
	}
	html := out.String()
	if !strings.Contains(html, "<table>") {
		t.Fatalf("expected table in rendered html, got: %s", html)
	}
	if !strings.Contains(html, "<td>b</td>") {
		t.Fatalf("expected table cell in rendered html, got: %s", html)
	}
}

func TestServeNoteDetailRendersPermalink(t *testing.T) {
	workspace := newTestWorkspace(t)
	id := "20260408153042"
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", id+"-linking-notes.md"), "---\nid: "+id+"\ntitle: Linking Notes\ntype: permanent\ncreated: 2026-04-08T15:30:42Z\ntags: [links]\n---\n\nBody text.\n")

	app := workspace.app(t)
	h, err := app.newServeHandler("127.0.0.1:8787")
	if err != nil {
		t.Fatalf("newServeHandler: %v", err)
	}

	req := httptest.NewRequest("GET", "/note/"+id, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "copy permalink") {
		t.Fatalf("expected copy permalink action in rendered detail html, got: %s", body)
	}
	if !strings.Contains(body, `window.location.origin + '/note/`+id+`'`) {
		t.Fatalf("expected copy permalink action to target note permalink, got: %s", body)
	}
	if !strings.Contains(body, `href="/?tag=links"`) {
		t.Fatalf("expected clickable tag permalink filter in rendered detail html, got: %s", body)
	}
	if !strings.Contains(body, `hx-get="/notes"`) {
		t.Fatalf("expected tag click to refresh notes list endpoint, got: %s", body)
	}
	if !strings.Contains(body, `hx-include="#filters"`) {
		t.Fatalf("expected tag click to include current filters form values, got: %s", body)
	}
	if !strings.Contains(body, `hx-vals='js:{tag: "links"}'`) {
		t.Fatalf("expected tag click to send explicit tag value, got: %s", body)
	}
	if strings.Index(body, "copy permalink") > strings.Index(body, "<h4>Backlinks</h4>") {
		t.Fatalf("expected copy permalink action to appear before backlinks section")
	}
	if !strings.Contains(body, "<!doctype html>") {
		t.Fatalf("expected direct permalink request to render full page shell")
	}
	if !strings.Contains(body, "<style>") {
		t.Fatalf("expected full page to include styling")
	}
}

func TestServeNoteDetailHTMXReturnsPartial(t *testing.T) {
	workspace := newTestWorkspace(t)
	id := "20260408153042"
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", id+"-linking-notes.md"), "---\nid: "+id+"\ntitle: Linking Notes\ntype: permanent\ncreated: 2026-04-08T15:30:42Z\ntags: [links]\n---\n\nBody text.\n")

	app := workspace.app(t)
	h, err := app.newServeHandler("127.0.0.1:8787")
	if err != nil {
		t.Fatalf("newServeHandler: %v", err)
	}

	req := httptest.NewRequest("GET", "/note/"+id, nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "<!doctype html>") {
		t.Fatalf("expected htmx request to return detail partial, got full page")
	}
	if !strings.Contains(body, "copy permalink") {
		t.Fatalf("expected copy permalink action in partial detail html, got: %s", body)
	}
	if !strings.Contains(body, `window.location.origin + '/note/`+id+`'`) {
		t.Fatalf("expected copy permalink action to target note permalink in partial, got: %s", body)
	}
}

func TestBuildFrontmatterKVDoesNotDuplicateTags(t *testing.T) {
	note := Note{
		Metadata: map[string]string{
			"id":    "202603051723",
			"title": "Final synthesis outline",
			"tags":  "[synthesis, final]",
		},
		Tags: []string{"synthesis", "final"},
	}

	kvs := buildFrontmatterKV(note)
	tagCount := 0
	for _, kv := range kvs {
		if kv.Key == "tags" {
			tagCount++
		}
	}

	if tagCount != 1 {
		t.Fatalf("expected one tags entry in frontmatter, got %d", tagCount)
	}
}

func TestBuildFrontmatterKVFallsBackToParsedTags(t *testing.T) {
	note := Note{
		Metadata: map[string]string{
			"id":    "202603051723",
			"title": "Final synthesis outline",
		},
		Tags: []string{"synthesis", "final"},
	}

	kvs := buildFrontmatterKV(note)
	for _, kv := range kvs {
		if kv.Key == "tags" {
			if kv.Value != "synthesis, final" {
				t.Fatalf("unexpected tags fallback value: %q", kv.Value)
			}
			return
		}
	}

	t.Fatalf("expected tags fallback entry in frontmatter")
}
