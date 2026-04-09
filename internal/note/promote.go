package note

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

func (app *App) runPromote(args []string) error {
	fs := flag.NewFlagSet("promote", flag.ContinueOnError)
	fs.SetOutput(nil)
	newTitle := fs.String("title", "", "optional replacement title for the promoted note")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 1 {
		return errors.New("usage: note promote [--title \"New Title\"] <id>")
	}

	path, err := app.promoteNote(fs.Args()[0], *newTitle)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func (app *App) promoteNote(id, newTitle string) (string, error) {
	path, note, _, err := app.findNoteByID(id)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(newTitle) != "" {
		note.Metadata["title"] = strings.TrimSpace(newTitle)
	}
	if note.Metadata["type"] == "permanent" {
		stampUpdated(note.Metadata)
		permanentCfg, ok := app.cfg.NoteTypes["permanent"]
		if !ok {
			return "", errors.New("note.json has no permanent note type")
		}
		newPath := app.notePath(filepath.Dir(path), note.Metadata["id"], note.Metadata["title"])
		if err := writeNoteFile(path, newPath, note, permanentCfg); err != nil {
			return "", err
		}
		app.invalidateIndex()
		return newPath, nil
	}
	if note.Metadata["type"] == "permanent" {
		return path, nil
	}

	permanentCfg, ok := app.cfg.NoteTypes["permanent"]
	if !ok {
		return "", errors.New("note.json has no permanent note type")
	}

	note.Metadata["type"] = "permanent"
	delete(note.Metadata, "project")
	stampUpdated(note.Metadata)

	targetDir := filepath.Join(app.root, filepath.FromSlash(permanentCfg.Dir))
	if err := ensureDirectory(targetDir); err != nil {
		return "", fmt.Errorf("create permanent note directory: %w", err)
	}
	newPath := app.notePath(targetDir, note.Metadata["id"], note.Metadata["title"])
	if strings.EqualFold(filepath.Clean(path), filepath.Clean(newPath)) {
		if err := writeNoteFile(path, path, note, permanentCfg); err != nil {
			return "", err
		}
		app.invalidateIndex()
		return path, nil
	}
	if err := writeNoteFile(path, newPath, note, permanentCfg); err != nil {
		return "", err
	}
	app.invalidateIndex()
	return newPath, nil
}
