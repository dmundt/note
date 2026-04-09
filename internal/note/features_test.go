package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── tags ─────────────────────────────────────────────────────────────────────

func TestRunTagsCountsOccurrences(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081900-a.md"), "---\nid: 202604081900\ntitle: Alpha\ntype: permanent\ncreated: 2026-04-08T19:00:00Z\ntags: [zettel, links]\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081901-b.md"), "---\nid: 202604081901\ntitle: Beta\ntype: permanent\ncreated: 2026-04-08T19:01:00Z\ntags: [zettel]\n---\nbody\n")

	app := workspace.app(t)
	counts, err := app.collectTags("")
	if err != nil {
		t.Fatalf("collectTags: %v", err)
	}
	if counts["zettel"] != 2 {
		t.Errorf("expected zettel count 2, got %d", counts["zettel"])
	}
	if counts["links"] != 1 {
		t.Errorf("expected links count 1, got %d", counts["links"])
	}
}

func TestRunTagsEmptyWorkspace(t *testing.T) {
	workspace := newTestWorkspace(t)
	app := workspace.app(t)
	counts, err := app.collectTags("")
	if err != nil {
		t.Fatalf("collectTags: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty counts, got %v", counts)
	}
}

func TestRunTagsFilterByType(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "202604081902-a.md"), "---\nid: 202604081902\ntitle: Perm\ntype: permanent\ncreated: 2026-04-08T19:02:00Z\ntags: [keeper]\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "202604081903-b.md"), "---\nid: 202604081903\ntitle: Fleet\ntype: fleeting\ncreated: 2026-04-08T19:03:00Z\ntags: [inbox]\n---\nbody\n")

	app := workspace.app(t)
	counts, err := app.collectTags("permanent")
	if err != nil {
		t.Fatalf("collectTags: %v", err)
	}
	if counts["keeper"] != 1 {
		t.Errorf("expected keeper=1, got %d", counts["keeper"])
	}
	if counts["inbox"] != 0 {
		t.Errorf("expected inbox=0 when filtered to permanent, got %d", counts["inbox"])
	}
}

// ── recent ────────────────────────────────────────────────────────────────────

func TestRunRecentDefaultsToTen(t *testing.T) {
	workspace := newTestWorkspace(t)
	// Write 12 notes; runRecent should return only the 10 most recent.
	ids := []string{
		"20260409190100", "20260409190200", "20260409190300", "20260409190400",
		"20260409190500", "20260409190600", "20260409190700", "20260409190800",
		"20260409190900", "20260409191000", "20260409191100", "20260409191200",
	}
	for _, id := range ids {
		writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", id+"-x.md"),
			"---\nid: "+id+"\ntitle: Note "+id+"\ntype: fleeting\ncreated: 2026-04-09T19:00:00Z\ntags: []\n---\nbody\n")
	}

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runRecent(nil)
	})
	if err != nil {
		t.Fatalf("runRecent: %v", err)
	}
	lines := nonEmptyLines(output)
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d:\n%s", len(lines), output)
	}
}

func TestRunRecentCustomCount(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260409200100-a.md"), "---\nid: 20260409200100\ntitle: A\ntype: permanent\ncreated: 2026-04-09T20:01:00Z\ntags: []\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260409200200-b.md"), "---\nid: 20260409200200\ntitle: B\ntype: permanent\ncreated: 2026-04-09T20:02:00Z\ntags: []\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260409200300-c.md"), "---\nid: 20260409200300\ntitle: C\ntype: permanent\ncreated: 2026-04-09T20:03:00Z\ntags: []\n---\nbody\n")

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runRecent([]string{"2"})
	})
	if err != nil {
		t.Fatalf("runRecent: %v", err)
	}
	lines := nonEmptyLines(output)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[0], "20260409200300") {
		t.Errorf("expected most recent note first, got: %s", lines[0])
	}
}

func TestRunRecentInvalidCount(t *testing.T) {
	workspace := newTestWorkspace(t)
	app := workspace.app(t)
	err := app.runRecent([]string{"notanumber"})
	if err == nil {
		t.Fatal("expected error for invalid count")
	}
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestRunListAllNotes(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260409210100-p.md"), "---\nid: 20260409210100\ntitle: Permanent One\ntype: permanent\ncreated: 2026-04-09T21:01:00Z\ntags: []\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "20260409210200-f.md"), "---\nid: 20260409210200\ntitle: Fleeting One\ntype: fleeting\ncreated: 2026-04-09T21:02:00Z\ntags: []\n---\nbody\n")

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runList(nil)
	})
	if err != nil {
		t.Fatalf("runList: %v", err)
	}
	if !strings.Contains(output, "20260409210100") {
		t.Errorf("expected permanent note in list: %s", output)
	}
	if !strings.Contains(output, "20260409210200") {
		t.Errorf("expected fleeting note in list: %s", output)
	}
}

func TestRunListFilterByType(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260409220100-p.md"), "---\nid: 20260409220100\ntitle: Perm\ntype: permanent\ncreated: 2026-04-09T22:01:00Z\ntags: []\n---\nbody\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "20260409220200-f.md"), "---\nid: 20260409220200\ntitle: Fleet\ntype: fleeting\ncreated: 2026-04-09T22:02:00Z\ntags: []\n---\nbody\n")

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runList([]string{"--type", "permanent"})
	})
	if err != nil {
		t.Fatalf("runList: %v", err)
	}
	if !strings.Contains(output, "20260409220100") {
		t.Errorf("expected permanent note: %s", output)
	}
	if strings.Contains(output, "20260409220200") {
		t.Errorf("expected fleeting note to be filtered out: %s", output)
	}
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestRunDeleteRemovesFile(t *testing.T) {
	workspace := newTestWorkspace(t)
	notePath := filepath.Join("notes", "fleeting", "20260409230100-todelete.md")
	writeTestNote(t, workspace.root, notePath, "---\nid: 20260409230100\ntitle: To Delete\ntype: fleeting\ncreated: 2026-04-09T23:01:00Z\ntags: []\n---\nbody\n")

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runDelete([]string{"20260409230100"})
	})
	if err != nil {
		t.Fatalf("runDelete: %v", err)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output: %s", output)
	}
	if fileExists(filepath.Join(workspace.root, notePath)) {
		t.Error("expected file to be removed")
	}
}

func TestRunDeleteErrorsOnUnknownID(t *testing.T) {
	workspace := newTestWorkspace(t)
	app := workspace.app(t)
	err := app.runDelete([]string{"99999999999999"})
	if err == nil {
		t.Fatal("expected error for unknown ID")
	}
}

// ── link ──────────────────────────────────────────────────────────────────────

func TestRunLinkInjectsLink(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410100100-from.md"), "---\nid: 20260410100100\ntitle: From Note\ntype: permanent\ncreated: 2026-04-10T10:01:00Z\ntags: []\n---\n\nSome content here.\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410100200-to.md"), "---\nid: 20260410100200\ntitle: To Note\ntype: permanent\ncreated: 2026-04-10T10:02:00Z\ntags: []\n---\n\nTarget content.\n")

	app := workspace.app(t)
	if _, err := captureStdout(t, func() error {
		return app.runLink([]string{"20260410100100", "20260410100200"})
	}); err != nil {
		t.Fatalf("runLink: %v", err)
	}

	fromNote, parseErr := parseNoteFile(filepath.Join(workspace.root, "notes", "permanent", "20260410100100-from.md"))
	if parseErr != nil {
		t.Fatalf("re-parse from note: %v", parseErr)
	}
	if !strings.Contains(fromNote.Body, "20260410100200-to.md") {
		t.Errorf("expected link in body, got:\n%s", fromNote.Body)
	}
	if !strings.Contains(fromNote.Body, "To Note") {
		t.Errorf("expected link title in body, got:\n%s", fromNote.Body)
	}
}

func TestRunLinkCreatesSectionIfAbsent(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410110100-from.md"), "---\nid: 20260410110100\ntitle: From\ntype: permanent\ncreated: 2026-04-10T11:01:00Z\ntags: []\n---\n\nNo links section yet.\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410110200-to.md"), "---\nid: 20260410110200\ntitle: To\ntype: permanent\ncreated: 2026-04-10T11:02:00Z\ntags: []\n---\n\nTarget.\n")

	app := workspace.app(t)
	if _, err := captureStdout(t, func() error {
		return app.runLink([]string{"20260410110100", "20260410110200"})
	}); err != nil {
		t.Fatalf("runLink: %v", err)
	}

	fromNote, _ := parseNoteFile(filepath.Join(workspace.root, "notes", "permanent", "20260410110100-from.md"))
	if !strings.Contains(fromNote.Body, "## Links") {
		t.Errorf("expected ## Links section to be created:\n%s", fromNote.Body)
	}
}

func TestRunLinkSkipsDuplicate(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410120100-from.md"), "---\nid: 20260410120100\ntitle: From\ntype: permanent\ncreated: 2026-04-10T12:01:00Z\ntags: []\n---\n\n## Links\n- [To](./20260410120200-to.md)\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410120200-to.md"), "---\nid: 20260410120200\ntitle: To\ntype: permanent\ncreated: 2026-04-10T12:02:00Z\ntags: []\n---\n\nTarget.\n")

	app := workspace.app(t)
	if err := app.runLink([]string{"20260410120100", "20260410120200"}); err != nil {
		t.Fatalf("runLink: %v", err)
	}
	fromNote, _ := parseNoteFile(filepath.Join(workspace.root, "notes", "permanent", "20260410120100-from.md"))
	count := strings.Count(fromNote.Body, "20260410120200-to.md")
	if count != 1 {
		t.Errorf("expected exactly 1 link, got %d:\n%s", count, fromNote.Body)
	}
}

// ── export ────────────────────────────────────────────────────────────────────

func TestExportContainsAllNotes(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410130100-a.md"), "---\nid: 20260410130100\ntitle: Alpha Export\ntype: permanent\ncreated: 2026-04-10T13:01:00Z\ntags: []\n---\n\nAlpha body.\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "20260410130200-b.md"), "---\nid: 20260410130200\ntitle: Beta Export\ntype: fleeting\ncreated: 2026-04-10T13:02:00Z\ntags: []\n---\n\nBeta body.\n")

	app := workspace.app(t)
	results, err := app.searchNotes("", "", "", "")
	if err != nil {
		t.Fatalf("searchNotes: %v", err)
	}
	out, err := buildExport(results)
	if err != nil {
		t.Fatalf("buildExport: %v", err)
	}
	if !strings.Contains(out, "# Alpha Export") {
		t.Errorf("expected Alpha Export heading: %s", out)
	}
	if !strings.Contains(out, "# Beta Export") {
		t.Errorf("expected Beta Export heading: %s", out)
	}
	if !strings.Contains(out, "Alpha body.") {
		t.Errorf("expected Alpha body: %s", out)
	}
}

func TestExportFilterByType(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, filepath.Join("notes", "permanent", "20260410140100-p.md"), "---\nid: 20260410140100\ntitle: Perm Export\ntype: permanent\ncreated: 2026-04-10T14:01:00Z\ntags: []\n---\n\nPerm body.\n")
	writeTestNote(t, workspace.root, filepath.Join("notes", "fleeting", "20260410140200-f.md"), "---\nid: 20260410140200\ntitle: Fleet Export\ntype: fleeting\ncreated: 2026-04-10T14:02:00Z\ntags: []\n---\n\nFleet body.\n")

	app := workspace.app(t)
	output, err := captureStdout(t, func() error {
		return app.runExport([]string{"--type", "permanent"})
	})
	if err != nil {
		t.Fatalf("runExport: %v", err)
	}
	if !strings.Contains(output, "Perm Export") {
		t.Errorf("expected Perm Export in output: %s", output)
	}
	if strings.Contains(output, "Fleet Export") {
		t.Errorf("unexpected Fleet Export in filtered output: %s", output)
	}
}

// ── completion ────────────────────────────────────────────────────────────────

func TestCompletionOutputsBash(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return runCompletion([]string{"bash"})
	})
	if err != nil {
		t.Fatalf("runCompletion bash: %v", err)
	}
	if !strings.Contains(output, "complete -F _note_completion note") {
		t.Errorf("bash completion missing register line: %s", output)
	}
}

func TestCompletionOutputsPowerShell(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return runCompletion([]string{"powershell"})
	})
	if err != nil {
		t.Fatalf("runCompletion powershell: %v", err)
	}
	if !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("powershell completion missing Register-ArgumentCompleter: %s", output)
	}
}

func TestCompletionRejectsUnknownShell(t *testing.T) {
	err := runCompletion([]string{"cmd"})
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

// ── injectLink ────────────────────────────────────────────────────────────────

func TestInjectLinkNoSection(t *testing.T) {
	body := "Some content.\n"
	result := injectLink(body, "- [Target](./target.md)")
	if !strings.Contains(result, "## Links") {
		t.Errorf("expected ## Links section: %s", result)
	}
	if !strings.Contains(result, "- [Target](./target.md)") {
		t.Errorf("expected link line: %s", result)
	}
}

func TestInjectLinkExistingSection(t *testing.T) {
	body := "Some content.\n\n## Links\n- [Old](./old.md)\n"
	result := injectLink(body, "- [New](./new.md)")
	if strings.Count(result, "## Links") != 1 {
		t.Errorf("expected exactly one ## Links section: %s", result)
	}
	if !strings.Contains(result, "- [Old](./old.md)") {
		t.Errorf("expected old link preserved: %s", result)
	}
	if !strings.Contains(result, "- [New](./new.md)") {
		t.Errorf("expected new link added: %s", result)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func nonEmptyLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
