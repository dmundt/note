package note

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const recentDefault = 10

func (app *App) runRecent(args []string) error {
	fs := flag.NewFlagSet("recent", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	noteType := fs.String("type", "", "filter by note type")
	if err := fs.Parse(args); err != nil {
		return err
	}

	n := recentDefault
	if len(fs.Args()) == 1 {
		var err error
		n, err = strconv.Atoi(fs.Args()[0])
		if err != nil || n <= 0 {
			return fmt.Errorf("recent: invalid count %q (must be a positive integer)", fs.Args()[0])
		}
	} else if len(fs.Args()) > 1 {
		return fmt.Errorf("usage: note recent [--type <type>] [n]")
	}

	idx, err := app.getIndex()
	if err != nil {
		return err
	}

	notes := idx.notes
	if *noteType != "" {
		if _, ok := app.cfg.NoteTypes[*noteType]; !ok {
			return fmt.Errorf("unknown note type %q", *noteType)
		}
		filtered := notes[:0:0]
		for _, note := range notes {
			if note.Metadata["type"] == *noteType {
				filtered = append(filtered, note)
			}
		}
		notes = filtered
	}

	if n > len(notes) {
		n = len(notes)
	}

	output, err := formatSearchResults(app.root, notes[:n], "pipe")
	if err != nil {
		return err
	}
	if output != "" {
		fmt.Print(output)
	}
	return nil
}
