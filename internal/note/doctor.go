package note

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

func (app *App) runDoctor(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: note doctor")
	}

	issues, checkedCount, err := app.collectDoctorIssues()
	if err != nil {
		return err
	}
	if len(issues) == 0 {
		fmt.Printf("doctor: ok (%d notes checked)\n", checkedCount)
		return nil
	}

	fmt.Printf("doctor: found %d issue(s)\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("- %s\n", issue)
	}
	return errors.New("doctor found issues")
}

func (app *App) collectDoctorIssues() ([]string, int, error) {
	idx, err := app.getIndex()
	if err != nil {
		return nil, 0, err
	}

	issues := make([]string, 0)
	for _, parseErr := range idx.parseErrors {
		issues = append(issues, "malformed note: "+parseErr)
	}
	idPaths := map[string][]string{}
	optionalFields := map[string]bool{"source": true, "updated": true}

	allowedDirs := make([]string, 0, len(app.cfg.NoteTypes))
	for _, typeCfg := range app.cfg.NoteTypes {
		allowedDirs = append(allowedDirs, filepath.Clean(filepath.Join(app.root, filepath.FromSlash(typeCfg.Dir))))
	}

	for _, note := range idx.notes {
		noteID := strings.TrimSpace(note.Metadata["id"])
		if noteID == "" {
			issues = append(issues, fmt.Sprintf("%s: missing id", relativeToRoot(app.root, note.Path)))
		} else {
			idPaths[noteID] = append(idPaths[noteID], relativeToRoot(app.root, note.Path))
		}

		noteType := strings.TrimSpace(note.Metadata["type"])
		typeCfg, ok := app.cfg.NoteTypes[noteType]
		if !ok {
			issues = append(issues, fmt.Sprintf("%s: unknown note type %q", relativeToRoot(app.root, note.Path), noteType))
			continue
		}

		for _, key := range typeCfg.Frontmatter {
			if key == "tags" {
				continue
			}
			if optionalFields[key] {
				continue
			}
			if strings.TrimSpace(note.Metadata[key]) == "" {
				issues = append(issues, fmt.Sprintf("%s: missing required field %q", relativeToRoot(app.root, note.Path), key))
			}
		}

		expectedDir := filepath.Clean(filepath.Join(app.root, filepath.FromSlash(typeCfg.Dir)))
		noteDir := filepath.Clean(filepath.Dir(note.Path))
		if noteDir != expectedDir {
			issues = append(issues, fmt.Sprintf("%s: note type %q is in unexpected directory %s", relativeToRoot(app.root, note.Path), noteType, relativeToRoot(app.root, noteDir)))
		}

		issues = append(issues, app.findBrokenMarkdownLinks(note)...)
	}

	for id, paths := range idPaths {
		if len(paths) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate id %s in %s", id, strings.Join(paths, ", ")))
		}
	}

	unknownDirIssues, err := app.findUnknownNoteDirectories(allowedDirs)
	if err != nil {
		return nil, 0, err
	}
	issues = append(issues, unknownDirIssues...)
	return issues, len(idx.notes), nil
}

func (app *App) findBrokenMarkdownLinks(note Note) []string {
	issues := []string{}
	matches := markdownLinkPattern.FindAllStringSubmatch(note.Body, -1)
	baseDir := filepath.Dir(note.Path)

	for _, match := range matches {
		target := strings.TrimSpace(match[1])
		if target == "" {
			continue
		}
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(target), ".md") {
			continue
		}

		resolved := target
		if !filepath.IsAbs(target) {
			resolved = filepath.Join(baseDir, filepath.FromSlash(target))
		}
		if _, err := os.Stat(resolved); err != nil {
			issues = append(issues, fmt.Sprintf("%s: broken markdown link %q", relativeToRoot(app.root, note.Path), target))
		}
	}
	return issues
}

func (app *App) findUnknownNoteDirectories(allowedDirs []string) ([]string, error) {
	notesRoot := filepath.Join(app.root, "notes")
	if _, err := os.Stat(notesRoot); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	issues := []string{}
	err := filepath.WalkDir(notesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		noteDir := filepath.Clean(filepath.Dir(path))
		for _, allowedDir := range allowedDirs {
			if noteDir == allowedDir {
				return nil
			}
		}
		issues = append(issues, fmt.Sprintf("%s: note file is in unknown notes directory", relativeToRoot(app.root, path)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return issues, nil
}
