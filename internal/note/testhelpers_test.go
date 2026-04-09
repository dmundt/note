package note

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type testWorkspace struct {
	root string
	cfg  Config
}

func newTestWorkspace(t *testing.T) testWorkspace {
	t.Helper()

	root := t.TempDir()
	cfg := Config{
		Version:             1,
		IDFormat:            "200601021504",
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

	writeJSON(t, filepath.Join(root, "note.json"), cfg)
	writeTestNote(t, root, filepath.Join("templates", "fleeting.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: fleeting\ncreated: {{created}}\ntags: []\nsource:\n---\n\n{{body}}\n")
	writeTestNote(t, root, filepath.Join("templates", "permanent.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: permanent\ncreated: {{created}}\ntags: []\n---\n\n{{body}}\n")
	writeTestNote(t, root, filepath.Join("templates", "project.md"), "---\nid: {{id}}\ntitle: {{title}}\ntype: project\nproject: {{project}}\ncreated: {{created}}\ntags: []\n---\n\n{{body}}\n")

	return testWorkspace{root: root, cfg: cfg}
}

func (workspace testWorkspace) app(t *testing.T) *App {
	t.Helper()
	return &App{root: workspace.root, cfg: workspace.cfg}
}

func (workspace testWorkspace) withWorkingDirectory(t *testing.T, run func()) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(workspace.root); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()
	run()
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	writeTestNote(t, filepath.Dir(path), filepath.Base(path), string(data))
}

func writeTestNote(t *testing.T, root, relativePath, content string) {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(data)
}

func assertGoldenFile(t *testing.T, relativePath, actual string) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test helper path")
	}
	goldenPath := filepath.Join(filepath.Dir(currentFile), relativePath)
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v", goldenPath, err)
	}
	want := strings.ReplaceAll(string(data), "\r\n", "\n")
	got := strings.ReplaceAll(actual, "\r\n", "\n")
	want = strings.TrimRight(want, "\n")
	got = strings.TrimRight(got, "\n")
	if want != got {
		t.Fatalf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", goldenPath, want, got)
	}
}

func normalizeDynamicNoteContent(content string) string {
	createdPattern := regexp.MustCompile(`(?m)^created: .*$`)
	updatedPattern := regexp.MustCompile(`(?m)^updated: .*$`)
	content = createdPattern.ReplaceAllString(content, "created: <created>")
	content = updatedPattern.ReplaceAllString(content, "updated: <updated>")
	return content
}
