# Stickies

Quick notes and a daily journal in one TUI. Pinnable stickies on the left, dated journal entries on the right, fast keyboard editing for both. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Current release: **v1.0.0**.

![stickies screenshot](docs/screenshot.png)

## Quick Install

Supported platforms: Linux and macOS. On Windows, use WSL.

Recommended (installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/LFroesch/stickies/main/install.sh | bash
```

Or download a binary from [GitHub Releases](https://github.com/LFroesch/stickies/releases).

Or install with Go:

```bash
go install github.com/LFroesch/stickies@latest
```

Or build from source:

```bash
make install
```

## Usage

A `stk` alias is symlinked alongside the `stickies` binary on install. The two are interchangeable.

```bash
stickies                       # launch the TUI
stk                            # same, shorter
stk --version
stk --help
stickies --data ./mydir        # override data directory

# Pipeable subcommands (no TUI):
stk ls                         # list stickies: id<TAB>pin<TAB>title
stk cat <id|title>             # print a sticky's body to stdout
stk new "buy milk"             # create sticky from arg (or pipe via stdin)
stk log "finished cli wiring"  # append timestamped line to today's journal
stk today                      # print today's journal entry
stk day 2026-04-01             # print a specific day
stk days                       # list all journal dates
stk search "todo"              # grep across stickies + journal

# Stdin works for new/log too:
echo "multi-line note" | stk new
git log --oneline -5 | stk log
```

Data lives in `~/.local/share/stickies/` by default (overridable via `$STICKIES_DIR` or `--data`). The `$EDITOR` env var is honored when editing notes externally (`o` key).
Quoted editor commands such as `EDITOR='code --wait'` are supported.

## Keybinds

| Key            | Action                              |
|----------------|-------------------------------------|
| `q`            | Quit                                |
| `?`            | Toggle help                         |
| `1` / `2` / `3`  | Stickies / Journal / Search page  |
| `tab` / `shift+tab` | Cycle pages                    |
| `j`/`k`, `↑/↓`   | Move selection                    |
| `g` / `G`        | Jump to top / bottom              |
| `ctrl+d`/`ctrl+u`| Page down / up                    |
| `n`              | New sticky                        |
| `t`              | Jump to / create today's journal  |
| `enter` / `E`    | Edit body (in-app)                |
| `e`              | Edit tags (stickies only)         |
| `o`              | Open in `$EDITOR`                 |
| `y`              | Yank/copy body to clipboard       |
| `p`              | Toggle pin (stickies only)        |
| `D`              | Delete (with confirm)             |
| `/`              | Search across stickies + journal  |
| `ctrl+s`         | Save body (in edit mode)          |
| `ctrl+d`         | Delete char forward (edit mode)   |
| `home` / `end`   | Line start / end (edit mode)      |
| `ctrl+a` / `ctrl+e` | Line start / end (edit mode)   |
| `alt+b` / `alt+f` | Word back / forward (edit mode)  |
| `alt+d` / `ctrl+w` | Delete word fwd / back (edit mode) |
| `esc`            | Cancel mode / close overlay       |
| `q` / `ctrl+c`   | Quit                              |

## Environment

The following environment variables are honored:

| Var / flag       | Effect                                              |
|------------------|-----------------------------------------------------|
| `$VISUAL`        | External editor for `o` (preferred over `$EDITOR`)  |
| `$EDITOR`        | External editor fallback (`vi` if unset)            |
| `$STICKIES_DIR`  | Override data directory                             |
| `--data <dir>`   | Override data directory (wins over env)             |

## License

GPL-3.0 — see [LICENSE](LICENSE).
