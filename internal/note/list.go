package note

import (
	"flag"
	"fmt"
	"os"
)

func (app *App) runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	noteType := fs.String("type", "", "filter by note type")
	outputFormat := fs.String("format", app.cfg.SearchOutputDefault, "output format: pipe, plain, table, json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: note list [--type <type>] [--format <format>]")
	}

	results, err := app.searchNotes(*noteType, "", "", "")
	if err != nil {
		return err
	}

	output, err := formatSearchResults(app.root, results, *outputFormat)
	if err != nil {
		return err
	}
	if output != "" {
		fmt.Print(output)
	}
	return nil
}
