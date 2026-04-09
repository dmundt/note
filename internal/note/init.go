package note

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed initdata/note.json initdata/templates/*.md
var initFS embed.FS

func runInit(dir string, args []string) error {
	full := false
	if len(args) > 1 {
		return errors.New("usage: note init [full|--full]")
	}
	if len(args) == 1 {
		switch args[0] {
		case "full", "--full":
			full = true
		default:
			return errors.New("usage: note init [full|--full]")
		}
	}

	configPath := filepath.Join(dir, "note.json")
	if _, err := os.Stat(configPath); err == nil {
		return errors.New("note.json already exists; workspace is already initialised")
	}

	dirs := []string{
		filepath.Join(dir, "notes", "fleeting"),
		filepath.Join(dir, "notes", "permanent"),
		filepath.Join(dir, "notes", "project"),
	}
	if full {
		dirs = append(dirs, filepath.Join(dir, "templates"))
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
		fmt.Printf("created  %s/\n", filepath.ToSlash(d))
	}

	if err := writeEmbeddedFile(dir, "initdata/note.json", "note.json"); err != nil {
		return err
	}

	if full {
		if err := writeEmbeddedFile(dir, "initdata/templates/fleeting.md", filepath.ToSlash(filepath.Join("templates", "fleeting.md"))); err != nil {
			return err
		}
		if err := writeEmbeddedFile(dir, "initdata/templates/permanent.md", filepath.ToSlash(filepath.Join("templates", "permanent.md"))); err != nil {
			return err
		}
		if err := writeEmbeddedFile(dir, "initdata/templates/project.md", filepath.ToSlash(filepath.Join("templates", "project.md"))); err != nil {
			return err
		}
	}

	fmt.Println("initialised note workspace — run 'note new fleeting \"first thought\"' to get started")
	return nil
}

func writeEmbeddedFile(dir, embeddedPath, relDest string) error {
	data, err := initFS.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", embeddedPath, err)
	}
	dest := filepath.Join(dir, filepath.FromSlash(relDest))
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	fmt.Printf("created  %s\n", filepath.ToSlash(dest))
	return nil
}
