package note

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func writeNoteFile(currentPath, newPath string, note Note, typeCfg NoteTypeConfig) error {
	content := buildFrontmatter(note, typeCfg) + note.Body
	if currentPath != newPath {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("cannot rename note to %s: file already exists", newPath)
		}
	}

	tempPath := filepath.Join(filepath.Dir(newPath), fmt.Sprintf(".note-tmp-%d.md", time.Now().UnixNano()))
	if err := os.WriteFile(tempPath, []byte(content), 0o644); err != nil {
		return err
	}
	cleanupTemp := func() {
		_ = os.Remove(tempPath)
	}

	if currentPath == newPath {
		backupPath := currentPath + ".bak"
		_ = os.Remove(backupPath)
		if err := os.Rename(currentPath, backupPath); err != nil {
			cleanupTemp()
			return err
		}
		if err := os.Rename(tempPath, newPath); err != nil {
			_ = os.Rename(backupPath, currentPath)
			cleanupTemp()
			return err
		}
		_ = os.Remove(backupPath)
		return nil
	}

	if err := os.Rename(tempPath, newPath); err != nil {
		cleanupTemp()
		return err
	}
	if currentPath != newPath {
		if err := os.Remove(currentPath); err != nil {
			return err
		}
	}
	return nil
}

func buildFrontmatter(note Note, typeCfg NoteTypeConfig) string {
	ordered := append([]string{}, typeCfg.Frontmatter...)
	seen := map[string]bool{}
	for _, key := range ordered {
		seen[key] = true
	}

	var extraKeys []string
	for key, value := range note.Metadata {
		if key == "tags" || value == "" || seen[key] {
			continue
		}
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)
	ordered = append(ordered, extraKeys...)

	var lines []string
	lines = append(lines, "---")
	for _, key := range ordered {
		if key == "tags" {
			lines = append(lines, "tags: "+formatTags(note.Tags))
			continue
		}
		value := note.Metadata[key]
		if value == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", key, value))
	}
	if !seen["tags"] {
		lines = append(lines, "tags: "+formatTags(note.Tags))
	}
	lines = append(lines, "---", "")
	return strings.Join(lines, "\n")
}

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	return "[" + strings.Join(tags, ", ") + "]"
}
