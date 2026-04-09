package note

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Version             int                       `json:"version"`
	IDFormat            string                    `json:"idFormat"`
	FilenamePattern     string                    `json:"filenamePattern"`
	LinkStyle           string                    `json:"linkStyle"`
	SearchOutputDefault string                    `json:"searchOutputDefault"`
	NoteTypes           map[string]NoteTypeConfig `json:"noteTypes"`
}

type NoteTypeConfig struct {
	Dir         string   `json:"dir"`
	Template    string   `json:"template"`
	Frontmatter []string `json:"frontmatter"`
}

func loadApp() (*App, error) {
	root, err := findWorkspaceRoot()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(root, "note.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read note.json: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse note.json: %w", err)
	}
	if cfg.IDFormat == "" {
		return nil, errors.New("note.json is missing idFormat")
	}
	if cfg.SearchOutputDefault == "" {
		cfg.SearchOutputDefault = "pipe"
	}
	if cfg.SearchOutputDefault != "pipe" && cfg.SearchOutputDefault != "plain" && cfg.SearchOutputDefault != "table" && cfg.SearchOutputDefault != "json" {
		return nil, fmt.Errorf("note.json has invalid searchOutputDefault %q", cfg.SearchOutputDefault)
	}
	if len(cfg.NoteTypes) == 0 {
		return nil, errors.New("note.json has no noteTypes")
	}

	return &App{root: root, cfg: cfg}, nil
}

func findWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		candidate := filepath.Join(current, "note.json")
		if _, err := os.Stat(candidate); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("could not find note.json in the current directory or its parents")
		}
		current = parent
	}
}
