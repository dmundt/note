package note

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type noteIndex struct {
	notes       []Note
	parseErrors []string
}

func (app *App) invalidateIndex() {
	app.idx = nil
}

func (app *App) getIndex() (*noteIndex, error) {
	if app.idx != nil {
		return app.idx, nil
	}
	idx, err := app.buildIndex()
	if err != nil {
		return nil, err
	}
	app.idx = idx
	return idx, nil
}

func (app *App) buildIndex() (*noteIndex, error) {
	var searchDirs []string
	for _, typeCfg := range app.cfg.NoteTypes {
		searchDirs = append(searchDirs, filepath.Join(app.root, filepath.FromSlash(typeCfg.Dir)))
	}

	var notes []Note
	var parseErrors []string
	for _, dir := range searchDirs {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}
			note, err := parseNoteFile(path)
			if err != nil {
				parseErrors = append(parseErrors, filepath.ToSlash(path)+": "+err.Error())
				return nil
			}
			if note.Metadata["id"] == "" {
				return nil
			}
			notes = append(notes, note)
			return nil
		})
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Metadata["id"] > notes[j].Metadata["id"]
	})
	return &noteIndex{notes: notes, parseErrors: parseErrors}, nil
}
