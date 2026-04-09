package note

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) runNew(args []string) error {
	path, err := app.createNote(args)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func (app *App) createNote(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("usage: note new fleeting \"title\" | note new permanent \"title\" | note new project <project-slug> \"title\"")
	}

	noteType := args[0]
	typeCfg, ok := app.cfg.NoteTypes[noteType]
	if !ok {
		return "", fmt.Errorf("unknown note type %q", noteType)
	}

	vars := map[string]string{
		"type":    noteType,
		"body":    "",
		"created": time.Now().Format(time.RFC3339),
	}

	var title string
	if noteType == "project" {
		if len(args) < 3 {
			return "", errors.New("usage: note new project <project-slug> \"title\"")
		}
		vars["project"] = args[1]
		title = strings.Join(args[2:], " ")
	} else {
		title = strings.Join(args[1:], " ")
	}
	if strings.TrimSpace(title) == "" {
		return "", errors.New("title must not be empty")
	}

	id := time.Now().Format(app.cfg.IDFormat)
	vars["id"] = id
	vars["title"] = title

	slug := slugify(title)
	filename := strings.ReplaceAll(app.cfg.FilenamePattern, "{id}", id)
	filename = strings.ReplaceAll(filename, "{slug}", slug)

	templatePath := filepath.Join(app.root, filepath.FromSlash(typeCfg.Template))
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		embeddedTemplatePath := "initdata/" + filepath.ToSlash(typeCfg.Template)
		templateData, err = initFS.ReadFile(embeddedTemplatePath)
		if err != nil {
			return "", fmt.Errorf("read template %s: %w", typeCfg.Template, err)
		}
	}

	content := string(templateData)
	for key, value := range vars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}

	targetDir := filepath.Join(app.root, filepath.FromSlash(typeCfg.Dir))
	if err := ensureDirectory(targetDir); err != nil {
		return "", fmt.Errorf("create note directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write note: %w", err)
	}
	app.invalidateIndex()

	return targetPath, nil
}

func ensureDirectory(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (app *App) notePath(dir, id, title string) string {
	filename := strings.ReplaceAll(app.cfg.FilenamePattern, "{id}", id)
	filename = strings.ReplaceAll(filename, "{slug}", slugify(title))
	return filepath.Join(dir, filename)
}
