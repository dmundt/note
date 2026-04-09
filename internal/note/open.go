package note

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func (app *App) runOpen(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: note open <id>")
	}

	path, _, _, err := app.findNoteByID(args[0])
	if err != nil {
		return err
	}

	return openInEditor(path)
}

// openInEditor opens the given path in the user's preferred editor.
// It respects $VISUAL then $EDITOR, then falls back to a platform default.
func openInEditor(path string) error {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	if editor != "" {
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Platform fallbacks.
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default: // linux and others
		for _, candidate := range []string{"xdg-open", "nano", "vi"} {
			if _, err := exec.LookPath(candidate); err == nil {
				cmd := exec.Command(candidate, path)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}
		}
		return fmt.Errorf("no editor found; set $EDITOR or $VISUAL")
	}
}
