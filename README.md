# Minimal PKM Workspace

This workspace is a plain-text note system built around markdown files and a small Go CLI named `note`.

## Layout

```text
notes/
  fleeting/    quick capture, inbox, disposable thoughts
  permanent/   durable Zettelkasten notes with stable IDs and links
  project/     active work notes tied to a project slug
templates/
  fleeting.md
  permanent.md
  project.md
note.json      workspace contract for the note CLI
```

## Design Rules

- All notes are regular markdown files.
- Every note gets a stable timestamp ID in `YYYYMMDDHHmmss` format.
- Filenames are `{id}-{slug}.md`.
- Permanent notes should link to other permanent notes using wiki links like `[[20260408153042]]` or markdown links when needed.
- Frontmatter stays small and explicit.
- Project notes are working notes, not a second knowledge base. When a project insight becomes durable, promote it into `notes/permanent/`.

## Note Types

### Fleeting notes

Use for quick capture, half-formed ideas, and inbox material.

- Directory: `notes/fleeting/`
- Filename: `YYYYMMDDHHmmss-slug.md`
- Lifecycle: create fast, review soon, either delete or promote

### Permanent notes

Use for durable ideas worth keeping and linking.

- Directory: `notes/permanent/`
- Filename: `YYYYMMDDHHmmss-slug.md`
- Lifecycle: evergreen, link to related notes, refine over time

### Project notes

Use for active work tied to a bounded effort.

- Directory: `notes/project/`
- Filename: `YYYYMMDDHHmmss-slug.md`
- Required metadata: `project`
- Lifecycle: temporary working context; extract reusable ideas into permanent notes

## ID Scheme

Use a 14-digit local timestamp:

- Format: `YYYYMMDDHHmmss`
- Example: `20260408153042`
- Purpose: globally unique enough for one-user note creation, sortable, readable, and stable for links

Slug rules:

- lowercase words
- hyphen-separated
- keep short and descriptive
- example filename: `20260408153042-linking-notes-by-id.md`

## Lightweight Frontmatter

Common fields across all notes:

```yaml
---
id: 20260408153042
title: Linking Notes by ID
type: permanent
created: 2026-04-08T15:30:00-05:00
tags: [zettel, links]
---
```

Optional fields:

- `updated`: only when the note CLI changes metadata or title
- `project`: required for project notes, omitted otherwise
- `source`: optional for fleeting notes when captured from a meeting, article, or person

## Recommended Workflow

1. Capture quickly into `notes/fleeting/`.
2. Review fleeting notes and either delete them, convert them into project notes, or rewrite them as permanent notes.
3. Keep permanent notes atomic: one durable idea per note.
4. Link permanent notes by ID whenever the ideas relate.
5. Keep project notes scoped to active work and promote durable insights out of them.

## Exact note Commands

The workspace expects the `note` CLI to support these commands:

```text
note init
note init full
note completion <bash|zsh|powershell|fish>

note new fleeting "title"
note new permanent "title"
note new project <project-slug> "title"

note list
note list --type permanent
note list --format table

note recent
note recent 5

note tags
note tags --type permanent

note search "query"
note search --type permanent "query"
note search --type project --project <project-slug> "query"
note search --tag <tag>
note search --format json "query"
note search --format table "query"
note search --strict "query"

note backlinks <id>

note open <id>

note delete <id>

note link <from-id> <to-id>

note doctor

note stats

note export
note export --type permanent
note export --output all-notes.md

note promote <id>
note promote --title "New Title" <id>

note meta set <id> title "New Title"
note meta set <id> project <project-slug>
note meta add-tag <id> <tag>
note meta remove-tag <id> <tag>
```

`note search` output defaults to `pipe`, but can be configured in `note.json`:

```json
{
  "searchOutputDefault": "pipe"
}
```

Allowed values: `pipe`, `plain`, `table`, `json`.

`note search --strict` fails if malformed note files are present and reports each parse error.

`note doctor` exits with a non-zero status when issues are found, so it can be used in CI checks.

`note doctor` validates:
- duplicate IDs
- missing required metadata fields
- broken markdown links
- notes in unknown directories
- malformed note files discovered during indexing

`note init` creates `note.json` plus note directories by default and uses embedded fallback templates if `templates/` files are missing.

Use `note init full` when you also want workspace template files scaffolded into `templates/`.

See `note.json` for the workspace contract and `templates/` for file shapes.

## CLI Implementation

This workspace now includes a small Go implementation of `note`.

## Install As a Go Tool

The module path is `github.com/dmundt/note`, but the executable lives in `cmd/note`, so users should install it with:

```text
go install github.com/dmundt/note/cmd/note@latest
```

After you publish `v1.0.0`, the versioned install command becomes:

```text
go install github.com/dmundt/note/cmd/note@v1.0.0
```

`go install github.com/dmundt/note@v1.0.0` will not work for this repository because the module root is not a `main` package.

## Release v1.0.0

Use a semantic version tag at the repository root:

```text
git tag v1.0.0
git push origin v1.0.0
```

Recommended release flow:

1. run `go test ./...`
2. ensure the default branch is pushed
3. create and push the `v1.0.0` tag
4. optionally create a GitHub Release for notes and prebuilt binaries

Because this is the first stable major version, no `/v2`-style module suffix is needed. Keep the module path as `github.com/dmundt/note` for all `v1.x.y` releases.

Code layout:

- `cmd/note/` contains the thin CLI entrypoint.
- `internal/note/` contains config loading, note creation, search, metadata edits, and frontmatter handling.

Build it from the workspace root:

```text
go build -o note.exe ./cmd/note
```

Or run commands directly during development:

```text
go run ./cmd/note init
go run ./cmd/note init full
go run ./cmd/note completion powershell | Out-String | Invoke-Expression
go run ./cmd/note new fleeting "capture idea"
go run ./cmd/note list
go run ./cmd/note recent 5
go run ./cmd/note tags
go run ./cmd/note search --type permanent "linking"
go run ./cmd/note search --format json "linking"
go run ./cmd/note search --strict "linking"
go run ./cmd/note backlinks 20260408153042
go run ./cmd/note open 20260408153042
go run ./cmd/note link 20260408153042 20260408160000
go run ./cmd/note delete 20260408153042
go run ./cmd/note doctor
go run ./cmd/note stats
go run ./cmd/note export --output all-notes.md
go run ./cmd/note promote 20260408153042
go run ./cmd/note promote --title "Permanent Insight" 20260408153042
go run ./cmd/note meta add-tag 20260408153042 zettel
```

Completion scripts support live note ID suggestions for commands like `backlinks`, `open`, `delete`, `link`, `meta`, and `promote`.

Example outputs:

```text
go run ./cmd/note tags
zettel (38)
cat-behavior (17)
links (12)
```

```text
go run ./cmd/note recent 3
202602191922 | project | Case study — Siamese behavioral profile | notes/project/202602191922-catbehavior-case-study-siamese.md
202602121534 | project | Pattern analysis draft | notes/project/202602121534-catbehavior-pattern-analysis.md
202602051047 | project | Midpoint project summary | notes/project/202602051047-catbehavior-midpoint-summary.md
```

```text
go run ./cmd/note search "siamese"
202602191922 | project | Case study — Siamese behavioral profile | notes/project/202602191922-catbehavior-case-study-siamese.md
```

```text
go run ./cmd/note backlinks 202602191922
202602121534 | project | Pattern analysis draft | notes/project/202602121534-catbehavior-pattern-analysis.md
202602051047 | project | Midpoint project summary | notes/project/202602051047-catbehavior-midpoint-summary.md
```

```text
go run ./cmd/note doctor
doctor: ok (50 notes checked)
```

```text
go run ./cmd/note export --output all-notes.md
exported 50 note(s) to all-notes.md
```

```text
go run ./cmd/note doctor
doctor: found 2 issue(s)
- notes/permanent/202604081701-broken-link.md: broken markdown link "./missing.md"
- malformed note: notes/permanent/202604081702-bad.md: parse 202604081702-bad.md: invalid frontmatter line "title malformed"
error: doctor found issues
```

Run the test suite from the workspace root:

```text
go test ./...
```

Additional developer checks:

```text
go test -race ./...
go test -bench . ./internal/note
go vet ./...
```

## Troubleshooting

### `could not find note.json in the current directory or its parents`

Cause: running `note` outside the PKM workspace root.

Fix: run commands from the repository root (the directory that contains `note.json`), or any child directory under it.

### `duplicate note id ...; run note doctor to inspect duplicates`

Cause: at least two files share the same `id` value in frontmatter.

Fix:
1. run `note doctor` and find duplicate-id entries.
2. keep one ID unchanged and update the duplicate note `id` + filename to match.
3. update any links that pointed to the old duplicate ID.

### `strict mode: found ... malformed note(s)`

Cause: one or more note files have invalid frontmatter syntax.

Fix:
1. run `note search --strict "query"` or `note doctor`.
2. repair malformed frontmatter lines and ensure the frontmatter block is closed with `---`.

### `doctor found issues`

Cause: one or more validation checks failed (broken links, unknown directories, malformed files, metadata issues).

Fix: read each listed issue and correct files in place, then rerun `note doctor` until you get `doctor: ok (...)`.
