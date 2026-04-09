package note

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func (app *App) runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	noteType := fs.String("type", "", "filter by note type")
	outputFile := fs.String("output", "", "write to file instead of stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: note export [--type <type>] [--output <file>]")
	}

	results, err := app.searchNotes(*noteType, "", "", "")
	if err != nil {
		return err
	}

	out, err := buildExport(results)
	if err != nil {
		return err
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(out), 0o644); err != nil {
			return fmt.Errorf("export: write %s: %w", *outputFile, err)
		}
		fmt.Fprintf(os.Stderr, "exported %d note(s) to %s\n", len(results), *outputFile)
		return nil
	}

	fmt.Print(out)
	return nil
}

func buildExport(notes []Note) (string, error) {
	var sb strings.Builder
	for i, note := range notes {
		title := note.Metadata["title"]
		if title == "" {
			title = note.Metadata["id"]
		}
		noteType := note.Metadata["type"]
		id := note.Metadata["id"]

		fmt.Fprintf(&sb, "# %s\n\n", title)
		if noteType != "" || id != "" {
			sb.WriteString("<!-- ")
			parts := []string{}
			if id != "" {
				parts = append(parts, "id: "+id)
			}
			if noteType != "" {
				parts = append(parts, "type: "+noteType)
			}
			sb.WriteString(strings.Join(parts, ", "))
			sb.WriteString(" -->\n\n")
		}
		body := strings.TrimSpace(note.Body)
		if body != "" {
			sb.WriteString(body)
			sb.WriteString("\n")
		}
		if i < len(notes)-1 {
			sb.WriteString("\n---\n\n")
		}
	}
	return sb.String(), nil
}
