package note

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var wikiLinkPattern = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]*)?\]\]`)
var markdownLinkTargetPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

func (app *App) runBacklinks(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: note backlinks <id>")
	}
	results, err := app.backlinks(args[0])
	if err != nil {
		return err
	}
	output, err := formatSearchResults(app.root, results, "pipe")
	if err != nil {
		return err
	}
	if output != "" {
		fmt.Print(output)
	}
	return nil
}

func (app *App) backlinks(id string) ([]Note, error) {
	path, note, _, err := app.findNoteByID(id)
	if err != nil {
		return nil, err
	}
	targetFilename := filepath.Base(path)
	targetPath := filepath.Clean(path)

	idx, err := app.getIndex()
	if err != nil {
		return nil, err
	}

	var results []Note
	for _, candidate := range idx.notes {
		if candidate.Path == note.Path {
			continue
		}
		if hasBacklinkTo(candidate, id, targetFilename, targetPath) {
			results = append(results, candidate)
		}
	}
	return results, nil
}

func hasBacklinkTo(candidate Note, targetID, targetFilename, targetPath string) bool {
	body := candidate.Body

	wikiMatches := wikiLinkPattern.FindAllStringSubmatch(body, -1)
	for _, match := range wikiMatches {
		if len(match) < 2 {
			continue
		}
		if strings.TrimSpace(match[1]) == targetID {
			return true
		}
	}

	mdMatches := markdownLinkTargetPattern.FindAllStringSubmatch(body, -1)
	for _, match := range mdMatches {
		if len(match) < 2 {
			continue
		}
		rawTarget := strings.TrimSpace(match[1])
		if rawTarget == "" {
			continue
		}
		if strings.HasPrefix(rawTarget, "http://") || strings.HasPrefix(rawTarget, "https://") || strings.HasPrefix(rawTarget, "mailto:") || strings.HasPrefix(rawTarget, "#") {
			continue
		}
		target := rawTarget
		if hash := strings.Index(target, "#"); hash >= 0 {
			target = target[:hash]
		}
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		if path.Base(filepath.ToSlash(target)) == targetFilename {
			return true
		}

		resolved := target
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(filepath.Dir(candidate.Path), filepath.FromSlash(resolved))
		}
		if filepath.Clean(resolved) == targetPath {
			return true
		}
	}

	return false
}
