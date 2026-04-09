package note

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) runMeta(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: note meta set <id> <field> <value> | note meta add-tag <id> <tag> | note meta remove-tag <id> <tag>")
	}

	switch args[0] {
	case "set":
		if len(args) < 4 {
			return errors.New("usage: note meta set <id> <field> <value>")
		}
		path, err := app.metaSet(args[1], args[2], strings.Join(args[3:], " "))
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	case "add-tag":
		if len(args) != 3 {
			return errors.New("usage: note meta add-tag <id> <tag>")
		}
		path, err := app.metaAddTag(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	case "remove-tag":
		if len(args) != 3 {
			return errors.New("usage: note meta remove-tag <id> <tag>")
		}
		path, err := app.metaRemoveTag(args[1], args[2])
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	default:
		return fmt.Errorf("unknown meta command %q", args[0])
	}
}

func (app *App) metaSet(id, field, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("value must not be empty")
	}
	if field == "id" || field == "type" || field == "created" {
		return "", fmt.Errorf("field %q is immutable", field)
	}

	path, note, typeCfg, err := app.findNoteByID(id)
	if err != nil {
		return "", err
	}

	note.Metadata[field] = value
	if field == "title" {
		note.Metadata["title"] = value
	}
	stampUpdated(note.Metadata)

	newPath := path
	if field == "title" {
		newPath = app.renamedPath(path, note.Metadata["id"], value)
	}

	if err := writeNoteFile(path, newPath, note, typeCfg); err != nil {
		return "", err
	}
	app.invalidateIndex()

	return newPath, nil
}

func (app *App) metaAddTag(id, tag string) (string, error) {
	path, note, typeCfg, err := app.findNoteByID(id)
	if err != nil {
		return "", err
	}
	if containsFold(note.Tags, tag) {
		return path, nil
	}
	note.Tags = append(note.Tags, tag)
	stampUpdated(note.Metadata)
	if err := writeNoteFile(path, path, note, typeCfg); err != nil {
		return "", err
	}
	app.invalidateIndex()
	return path, nil
}

func (app *App) metaRemoveTag(id, tag string) (string, error) {
	path, note, typeCfg, err := app.findNoteByID(id)
	if err != nil {
		return "", err
	}
	filtered := note.Tags[:0]
	for _, existing := range note.Tags {
		if !strings.EqualFold(existing, tag) {
			filtered = append(filtered, existing)
		}
	}
	note.Tags = filtered
	stampUpdated(note.Metadata)
	if err := writeNoteFile(path, path, note, typeCfg); err != nil {
		return "", err
	}
	app.invalidateIndex()
	return path, nil
}

func (app *App) findNoteByID(id string) (string, Note, NoteTypeConfig, error) {
	idx, err := app.getIndex()
	if err != nil {
		return "", Note{}, NoteTypeConfig{}, err
	}

	var matches []Note
	for _, note := range idx.notes {
		if note.Metadata["id"] == id {
			matches = append(matches, note)
		}
	}

	if len(matches) > 1 {
		return "", Note{}, NoteTypeConfig{}, fmt.Errorf("duplicate note id %s; run `note doctor` to inspect duplicates", id)
	}
	if len(matches) == 1 {
		note := matches[0]
		typeCfg := app.cfg.NoteTypes[note.Metadata["type"]]
		return note.Path, note, typeCfg, nil
	}

	return "", Note{}, NoteTypeConfig{}, fmt.Errorf("could not find note with id %s", id)
}

func (app *App) renamedPath(currentPath, id, title string) string {
	dir := filepath.Dir(currentPath)
	filename := strings.ReplaceAll(app.cfg.FilenamePattern, "{id}", id)
	filename = strings.ReplaceAll(filename, "{slug}", slugify(title))
	return filepath.Join(dir, filename)
}

func stampUpdated(metadata map[string]string) {
	metadata["updated"] = time.Now().Format(time.RFC3339)
}
