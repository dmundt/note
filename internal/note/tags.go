package note

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func (app *App) runTags(args []string) error {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	noteType := fs.String("type", "", "filter by note type")
	if err := fs.Parse(args); err != nil {
		return err
	}

	counts, err := app.collectTags(*noteType)
	if err != nil {
		return err
	}

	if len(counts) == 0 {
		return nil
	}

	// Sort by count desc, then alpha.
	type tagCount struct {
		tag   string
		count int
	}
	ranked := make([]tagCount, 0, len(counts))
	for tag, count := range counts {
		ranked = append(ranked, tagCount{tag, count})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].count != ranked[j].count {
			return ranked[i].count > ranked[j].count
		}
		return ranked[i].tag < ranked[j].tag
	})

	var sb strings.Builder
	for _, tc := range ranked {
		fmt.Fprintf(&sb, "%s (%d)\n", tc.tag, tc.count)
	}
	fmt.Print(sb.String())
	return nil
}

func (app *App) collectTags(noteType string) (map[string]int, error) {
	notes, err := app.searchNotes(noteType, "", "", "")
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, note := range notes {
		for _, tag := range note.Tags {
			if tag != "" {
				counts[tag]++
			}
		}
	}
	return counts, nil
}
