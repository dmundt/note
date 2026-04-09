package note

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func (app *App) runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	noteType := fs.String("type", "", "filter by note type")
	project := fs.String("project", "", "filter project notes by project slug")
	tag := fs.String("tag", "", "filter by tag")
	outputFormat := fs.String("format", app.cfg.SearchOutputDefault, "output format: pipe, plain, table, json")
	strict := fs.Bool("strict", false, "fail if malformed notes are found while indexing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	results, err := app.searchNotes(*noteType, *project, *tag, query)
	if err != nil {
		return err
	}
	if *strict {
		idx, err := app.getIndex()
		if err != nil {
			return err
		}
		if len(idx.parseErrors) > 0 {
			return fmt.Errorf("strict mode: found %d malformed note(s):\n%s", len(idx.parseErrors), strings.Join(idx.parseErrors, "\n"))
		}
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

func formatSearchResults(root string, results []Note, outputFormat string) (string, error) {
	switch outputFormat {
	case "pipe", "plain":
		if len(results) == 0 {
			return "", nil
		}
		var builder strings.Builder
		for _, note := range results {
			fmt.Fprintf(&builder, "%s | %s | %s | %s\n",
				note.Metadata["id"],
				note.Metadata["type"],
				note.Metadata["title"],
				relativeToRoot(root, note.Path),
			)
		}
		return builder.String(), nil
	case "table":
		if len(results) == 0 {
			return "", nil
		}
		idWidth := len("id")
		typeWidth := len("type")
		titleWidth := len("title")
		for _, note := range results {
			if l := len(note.Metadata["id"]); l > idWidth {
				idWidth = l
			}
			if l := len(note.Metadata["type"]); l > typeWidth {
				typeWidth = l
			}
			if l := len(note.Metadata["title"]); l > titleWidth {
				titleWidth = l
			}
		}

		var builder strings.Builder
		fmt.Fprintf(&builder, "%-*s  %-*s  %-*s  %s\n", idWidth, "id", typeWidth, "type", titleWidth, "title", "path")
		for _, note := range results {
			fmt.Fprintf(&builder, "%-*s  %-*s  %-*s  %s\n",
				idWidth, note.Metadata["id"],
				typeWidth, note.Metadata["type"],
				titleWidth, note.Metadata["title"],
				relativeToRoot(root, note.Path),
			)
		}
		return builder.String(), nil
	case "json":
		if len(results) == 0 {
			return "[]\n", nil
		}
		rows := make([]map[string]string, 0, len(results))
		for _, note := range results {
			rows = append(rows, map[string]string{
				"id":    note.Metadata["id"],
				"type":  note.Metadata["type"],
				"title": note.Metadata["title"],
				"path":  relativeToRoot(root, note.Path),
			})
		}
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	default:
		return "", fmt.Errorf("unknown search output format %q (expected: pipe, plain, table, json)", outputFormat)
	}
}

func (app *App) searchNotes(noteType, project, tag, query string) ([]Note, error) {
	if noteType != "" {
		if _, ok := app.cfg.NoteTypes[noteType]; !ok {
			return nil, fmt.Errorf("unknown note type %q", noteType)
		}
	}

	idx, err := app.getIndex()
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	lowerQuery := strings.ToLower(query)
	lowerTag := strings.ToLower(tag)

	var results []Note
	for _, note := range idx.notes {
		if noteType != "" && note.Metadata["type"] != noteType {
			continue
		}
		if project != "" && note.Metadata["project"] != project {
			continue
		}
		if lowerTag != "" && !containsFold(note.Tags, lowerTag) {
			continue
		}
		if lowerQuery != "" {
			builder.Reset()
			builder.WriteString(note.Metadata["title"])
			builder.WriteByte('\n')
			builder.WriteString(note.Body)
			builder.WriteByte('\n')
			builder.WriteString(relativeToRoot(app.root, note.Path))
			haystack := strings.ToLower(builder.String())
			if !strings.Contains(haystack, lowerQuery) {
				continue
			}
		}
		results = append(results, note)
	}
	return results, nil
}
