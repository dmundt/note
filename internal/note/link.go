package note

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runLink injects a markdown link from the "from" note to the "to" note.
// It appends the link under a "## Links" section, creating the section if absent.
func (app *App) runLink(args []string) error {
	if len(args) != 2 {
		return errors.New("usage: note link <from-id> <to-id>")
	}
	fromID, toID := args[0], args[1]

	fromPath, fromNote, fromTypeCfg, err := app.findNoteByID(fromID)
	if err != nil {
		return fmt.Errorf("from note: %w", err)
	}
	toPath, toNote, _, err := app.findNoteByID(toID)
	if err != nil {
		return fmt.Errorf("to note: %w", err)
	}

	// Build a relative path from the from-note's directory to the to-note.
	fromDir := filepath.Dir(fromPath)
	rel, err := filepath.Rel(fromDir, toPath)
	if err != nil {
		return fmt.Errorf("compute relative path: %w", err)
	}
	linkTarget := filepath.ToSlash(rel)
	toTitle := toNote.Metadata["title"]
	if toTitle == "" {
		toTitle = toNote.Metadata["id"]
	}
	mdLink := fmt.Sprintf("- [%s](%s)", toTitle, linkTarget)

	// Check if the link already exists in the body.
	if strings.Contains(fromNote.Body, linkTarget) {
		fmt.Fprintf(os.Stderr, "note link: link to %s already present in %s\n", toID, fromID)
		return nil
	}

	fromNote.Body = injectLink(fromNote.Body, mdLink)

	if err := writeNoteFile(fromPath, fromPath, fromNote, fromTypeCfg); err != nil {
		return err
	}
	app.invalidateIndex()

	fmt.Printf("linked   %s → %s\n", relativeToRoot(app.root, fromPath), relativeToRoot(app.root, toPath))
	return nil
}

// injectLink appends mdLink under a "## Links" section.
// If the section exists, the link is appended before the next heading or EOF.
// If it doesn't exist, a new section is appended at the end of the body.
func injectLink(body, mdLink string) string {
	const section = "## Links"
	lines := strings.Split(body, "\n")

	// Find existing Links section.
	sectionIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == section {
			sectionIdx = i
			break
		}
	}

	if sectionIdx == -1 {
		// Append new section.
		trimmed := strings.TrimRight(body, "\n")
		return trimmed + "\n\n" + section + "\n" + mdLink + "\n"
	}

	// Find where to insert: right before the next ## heading or at end.
	insertAt := len(lines)
	for i := sectionIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") || strings.HasPrefix(lines[i], "# ") {
			insertAt = i
			break
		}
	}

	// Insert before insertAt.
	newLines := make([]string, 0, len(lines)+2)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, mdLink)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}
