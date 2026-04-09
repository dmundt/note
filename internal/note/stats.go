package note

import "fmt"

type NoteStats struct {
	Total     int
	Fleeting  int
	Permanent int
	Project   int
}

func (app *App) runStats(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: note stats")
	}

	stats, err := app.collectStats()
	if err != nil {
		return err
	}
	fmt.Print(formatStats(stats))
	return nil
}

func (app *App) collectStats() (NoteStats, error) {
	results, err := app.searchNotes("", "", "", "")
	if err != nil {
		return NoteStats{}, err
	}

	stats := NoteStats{Total: len(results)}
	for _, note := range results {
		switch note.Metadata["type"] {
		case "fleeting":
			stats.Fleeting++
		case "permanent":
			stats.Permanent++
		case "project":
			stats.Project++
		}
	}
	return stats, nil
}

func formatStats(stats NoteStats) string {
	return fmt.Sprintf("total: %d\nfleeting: %d\npermanent: %d\nproject: %d\n", stats.Total, stats.Fleeting, stats.Permanent, stats.Project)
}
