package note

import (
	"errors"
	"fmt"
)

const allCommands = "backlinks completion delete doctor export help init link list meta new open recent search serve stats tags promote"

func runCompletion(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: note completion <bash|zsh|powershell|fish>")
	}

	switch args[0] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "powershell":
		fmt.Print(powershellCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		return fmt.Errorf("unknown shell %q (supported: bash, zsh, powershell, fish)", args[0])
	}
	return nil
}

const bashCompletion = `# note bash completion
# Add to ~/.bashrc:  source <(note completion bash)
_note_completion() {
    local cur prev subcmd commands ids
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    subcmd="${COMP_WORDS[1]}"
    commands="` + allCommands + `"
    ids="$(note __ids 2>/dev/null)"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
        return 0
    fi

    case "${subcmd}" in
        backlinks|open|delete)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${ids}" -- "${cur}"))
                return 0
            fi ;;
        link)
            if [[ ${COMP_CWORD} -eq 2 || ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "${ids}" -- "${cur}"))
                return 0
            fi ;;
        meta)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "set add-tag remove-tag" -- "${cur}"))
                return 0
            fi
            if [[ ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "${ids}" -- "${cur}"))
                return 0
            fi ;;
        promote)
            if [[ "${prev}" == "--title" ]]; then
                return 0
            fi
            if [[ "${cur}" != -* ]]; then
                COMPREPLY=($(compgen -W "${ids}" -- "${cur}"))
                return 0
            fi ;;
    esac

    case "${prev}" in
        --type)
            COMPREPLY=($(compgen -W "fleeting permanent project" -- "${cur}"))
            return 0 ;;
        --format)
            COMPREPLY=($(compgen -W "pipe plain table json" -- "${cur}"))
            return 0 ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh powershell fish" -- "${cur}"))
            return 0 ;;
    esac
}
complete -F _note_completion note
`

const zshCompletion = `# note zsh completion
# Add to ~/.zshrc:  source <(note completion zsh)
_note() {
    local state
    local commands=(` + allCommands + `)
    local -a ids
    ids=(${(f)"$(note __ids 2>/dev/null)"})
    _arguments \
        '1: :->command' \
        '*: :->args'

    case $state in
        command) _describe 'command' commands ;;
        args)
            case ${words[2]} in
                backlinks|open|delete)
                    if (( CURRENT == 3 )); then
                        _describe 'note id' ids
                    fi ;;
                link)
                    if (( CURRENT == 3 || CURRENT == 4 )); then
                        _describe 'note id' ids
                    fi ;;
                meta)
                    if (( CURRENT == 3 )); then
                        _values 'meta command' set add-tag remove-tag
                    elif (( CURRENT == 4 )); then
                        _describe 'note id' ids
                    fi ;;
                promote)
                    _describe 'note id' ids ;;
                completion) _values 'shell' bash zsh powershell fish ;;
                search|list|recent|tags|export) _values 'flag' --type --format --tag --project --strict --output ;;
            esac ;;
    esac
}
compdef _note note
`

const powershellCompletion = `# note PowerShell completion
# Add to $PROFILE:  note completion powershell | Out-String | Invoke-Expression
Register-ArgumentCompleter -Native -CommandName note -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @('` + "backlinks', 'completion', 'delete', 'doctor', 'export', 'help', 'init', 'link', 'list', 'meta', 'new', 'open', 'recent', 'search', 'stats', 'tags', 'promote'" + `')
    $elements = $commandAst.CommandElements
    $subcommand = if ($elements.Count -gt 1) { $elements[1].ToString() } else { '' }

    if ($elements.Count -le 2) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }

    $ids = @(& note __ids 2>$null)
    $position = $elements.Count - 1

    if ($subcommand -in @('backlinks','open','delete') -and $position -eq 2) {
        $ids | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }
    if ($subcommand -eq 'link' -and $position -in @(2,3)) {
        $ids | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }
    if ($subcommand -eq 'meta') {
        if ($position -eq 2) {
            @('set','add-tag','remove-tag') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
            return
        }
        if ($position -eq 3) {
            $ids | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
            return
        }
    }
    if ($subcommand -eq 'promote' -and $wordToComplete -notlike '-*') {
        $ids | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }

    switch ($subcommand) {
        { $_ -in 'search','list','recent','tags' } {
            @('--type', '--format', '--tag', '--project', '--strict') |
                Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
                }
        }
        'completion' {
            @('bash','zsh','powershell','fish') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
        'export' {
            @('--type', '--output') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
    }
}
`

const fishCompletion = `# note fish completion
# Add to ~/.config/fish/completions/note.fish  (or: note completion fish > ~/.config/fish/completions/note.fish)
set -l commands ` + allCommands + `
function __note_ids
    note __ids 2>/dev/null
end
complete -c note -f -n '__fish_use_subcommand' -a "$commands"
complete -c note -n '__fish_seen_subcommand_from search list recent tags' -l type -a 'fleeting permanent project'
complete -c note -n '__fish_seen_subcommand_from search list' -l format -a 'pipe plain table json'
complete -c note -n '__fish_seen_subcommand_from completion' -a 'bash zsh powershell fish'
complete -c note -n '__fish_seen_subcommand_from export' -l type -a 'fleeting permanent project'
complete -c note -n '__fish_seen_subcommand_from export' -l output
complete -c note -n '__fish_seen_subcommand_from backlinks open delete' -a '(__note_ids)'
complete -c note -n '__fish_seen_subcommand_from link' -a '(__note_ids)'
complete -c note -n '__fish_seen_subcommand_from promote' -a '(__note_ids)'
complete -c note -n '__fish_seen_subcommand_from meta; and test (count (commandline -opc)) -eq 3' -a 'set add-tag remove-tag'
complete -c note -n '__fish_seen_subcommand_from meta; and test (count (commandline -opc)) -eq 4' -a '(__note_ids)'
`
