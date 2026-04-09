package note

import (
	"fmt"
)

// runIDs prints known note IDs (one per line) for shell completion.
func (app *App) runIDs(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: note __ids")
	}

	idx, err := app.getIndex()
	if err != nil {
		return err
	}

	seen := map[string]bool{}
	for _, note := range idx.notes {
		id := note.Metadata["id"]
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		fmt.Println(id)
	}
	return nil
}
