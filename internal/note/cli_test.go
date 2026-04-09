package note

import (
	"strings"
	"testing"
)

func TestRunCommandValidation(t *testing.T) {
	workspace := newTestWorkspace(t)
	writeTestNote(t, workspace.root, "notes/fleeting/202604081603-smoke-test-capture.md", `---
id: 202604081603
title: smoke test capture
type: fleeting
created: 2026-04-08T16:03:18Z
tags: [inbox]
source:
---

`)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "unknown command",
			args:    []string{"bogus"},
			wantErr: `unknown command "bogus"`,
		},
		{
			name:    "new missing args",
			args:    []string{"new"},
			wantErr: `usage: note new fleeting "title" | note new permanent "title" | note new project <project-slug> "title"`,
		},
		{
			name:    "new unknown type",
			args:    []string{"new", "mystery", "title"},
			wantErr: `unknown note type "mystery"`,
		},
		{
			name:    "new missing project title",
			args:    []string{"new", "project", "sf1"},
			wantErr: `usage: note new project <project-slug> "title"`,
		},
		{
			name:    "search unknown type",
			args:    []string{"search", "--type", "mystery"},
			wantErr: `unknown note type "mystery"`,
		},
		{
			name:    "search unknown format",
			args:    []string{"search", "--format", "xml", "query"},
			wantErr: `unknown search output format "xml"`,
		},
		{
			name:    "backlinks missing args",
			args:    []string{"backlinks"},
			wantErr: `usage: note backlinks <id>`,
		},
		{
			name:    "stats unexpected args",
			args:    []string{"stats", "extra"},
			wantErr: `usage: note stats`,
		},
		{
			name:    "doctor unexpected args",
			args:    []string{"doctor", "extra"},
			wantErr: `usage: note doctor`,
		},
		{
			name:    "promote missing args",
			args:    []string{"promote"},
			wantErr: `usage: note promote [--title "New Title"] <id>`,
		},
		{
			name:    "promote missing id after title flag",
			args:    []string{"promote", "--title", "New Title"},
			wantErr: `usage: note promote [--title "New Title"] <id>`,
		},
		{
			name:    "promote missing note",
			args:    []string{"promote", "999999999999"},
			wantErr: `could not find note with id 999999999999`,
		},
		{
			name:    "meta missing args",
			args:    []string{"meta"},
			wantErr: `usage: note meta set <id> <field> <value> | note meta add-tag <id> <tag> | note meta remove-tag <id> <tag>`,
		},
		{
			name:    "meta missing set value",
			args:    []string{"meta", "set", "202604081603", "title"},
			wantErr: `usage: note meta set <id> <field> <value>`,
		},
		{
			name:    "meta immutable field",
			args:    []string{"meta", "set", "202604081603", "created", "x"},
			wantErr: `field "created" is immutable`,
		},
		{
			name:    "meta unknown subcommand",
			args:    []string{"meta", "bogus"},
			wantErr: `unknown meta command "bogus"`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			workspace.withWorkingDirectory(t, func() {
				err := Run(test.args)
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("unexpected error %q, want substring %q", err.Error(), test.wantErr)
				}
			})
		})
	}
}
