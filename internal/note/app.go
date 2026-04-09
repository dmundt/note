package note

import (
	"fmt"
	"path/filepath"
	"strings"
)

type App struct {
	root string
	cfg  Config
	idx  *noteIndex
}

type Note struct {
	Path     string
	Metadata map[string]string
	Tags     []string
	Body     string
}

func Run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usageText())
		return nil
	}

	// init and completion do not require an existing workspace.
	if args[0] == "init" {
		return runInit(".", args[1:])
	}
	if args[0] == "completion" {
		return runCompletion(args[1:])
	}

	app, err := loadApp()
	if err != nil {
		return err
	}

	switch args[0] {
	case "new":
		return app.runNew(args[1:])
	case "search":
		return app.runSearch(args[1:])
	case "serve":
		return app.runServe(args[1:])
	case "stats":
		return app.runStats(args[1:])
	case "backlinks":
		return app.runBacklinks(args[1:])
	case "doctor":
		return app.runDoctor(args[1:])
	case "promote":
		return app.runPromote(args[1:])
	case "meta":
		return app.runMeta(args[1:])
	case "list":
		return app.runList(args[1:])
	case "recent":
		return app.runRecent(args[1:])
	case "tags":
		return app.runTags(args[1:])
	case "help", "--help", "-h":
		fmt.Print(usageText())
		return nil
	case "open":
		return app.runOpen(args[1:])
	case "delete":
		return app.runDelete(args[1:])
	case "link":
		return app.runLink(args[1:])
	case "export":
		return app.runExport(args[1:])
	case "__ids":
		return app.runIDs(args[1:])
	case "init", "completion":
		// handled above; unreachable but keeps the switch exhaustive
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func relativeToRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var builder strings.Builder
	lastDash := false
	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "note"
	}
	return slug
}

func usageText() string {
	return `note manages plain markdown notes in the current workspace.

Usage:
	note help
	note init [full|--full]
	note completion <bash|zsh|powershell|fish>
	note new fleeting "title"
	note new permanent "title"
	note new project <project-slug> "title"
	note list [--type <type>] [--format <pipe|plain|table|json>]
	note recent [--type <type>] [n]
	note tags [--type <type>]
	note search [--type <type>] [--project <project-slug>] [--tag <tag>] [--format <pipe|plain|table|json>] [--strict] [query]
	note serve [--addr <host:port>] [--open]
	note backlinks <id>
	note open <id>
	note delete <id>
	note link <from-id> <to-id>
	note doctor
	note stats
	note export [--type <type>] [--output <file>]
	note promote [--title "New Title"] <id>
	note meta set <id> <field> <value>
	note meta add-tag <id> <tag>
	note meta remove-tag <id> <tag>

Commands:
	init        Scaffold a new note workspace. Default: notes + note.json; full: also create templates/.
	completion  Print shell completion script for bash, zsh, powershell, or fish.
	new         Create a note from the configured template and print the created file path.
	list        List all notes sorted by ID descending.
	recent      Show the N most recently created notes (default 10).
	tags        List all tags with occurrence counts.
	search      Search note titles, bodies, and paths; prints pipe-separated matches.
	serve       Start a local web UI (Go templates + HTMX) to browse rendered notes.
	backlinks   Show notes that link to a target note ID.
	open        Open a note in $VISUAL / $EDITOR.
	delete      Remove a note file and warn about dangling backlinks.
	link        Inject a markdown link from one note into another.
	doctor      Validate notes for duplicate IDs, metadata issues, and broken links.
	stats       Print note counts by type and a total count.
	export      Concatenate notes into a single Markdown document.
	promote     Move a fleeting or project note into permanent notes.
	meta        Update title, project, or tags for an existing note.

Examples:
	note init
	note init full
	note new fleeting "capture idea"
	note new project pkm "outline note flow"
	note list
	note list --type permanent
	note recent 5
	note tags
	note search --type permanent links
	note serve --open
	note backlinks 20260408153042
	note open 20260408153042
	note delete 20260408153042
	note link 20260408153042 20260408160000
	note doctor
	note stats
	note export --output all-notes.md
	note promote --title "Permanent Insight" 20260408153042
	note search --tag inbox
	note meta set 20260408153042 title "Better Title"
	note meta add-tag 20260408153042 zettel
	note completion powershell | Out-String | Invoke-Expression
`
}
