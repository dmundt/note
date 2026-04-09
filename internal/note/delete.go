package note

import (
	"errors"
	"fmt"
	"os"
)

func (app *App) runDelete(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: note delete <id>")
	}

	path, note, _, err := app.findNoteByID(args[0])
	if err != nil {
		return err
	}

	// Warn if other notes link to this one.
	bl, err := app.backlinks(args[0])
	if err != nil {
		return err
	}
	if len(bl) > 0 {
		fmt.Fprintf(os.Stderr, "warning: %d note(s) link to %s:\n", len(bl), note.Metadata["id"])
		for _, b := range bl {
			fmt.Fprintf(os.Stderr, "  %s | %s\n", b.Metadata["id"], b.Metadata["title"])
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	app.invalidateIndex()

	fmt.Printf("deleted  %s\n", relativeToRoot(app.root, path))
	return nil
}
