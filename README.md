# Stickies

Quick notes and a daily journal in one TUI. Pinnable stickies on the left, dated journal entries on the right, fast keyboard editing for both. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Current release: **v1.0.0**.

![stickies screenshot](docs/screenshot.png)

## Install

Supported platforms: Linux and macOS. On Windows, use WSL.

Recommended (installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/LFroesch/stickies/main/install.sh | bash
```

Or download a binary from [GitHub Releases](https://github.com/LFroesch/stickies/releases).

Or build from source:

```bash
make install
```

## Usage

```bash
stickies                       # launch the TUI
stickies --version
stickies --data ./mydir        # override data directory

# Pipeable subcommands (no TUI):
stickies ls                    # list stickies: id<TAB>pin<TAB>title
stickies cat <id|title>        # print a sticky's body to stdout
stickies today                 # print today's journal entry
stickies day 2026-04-01        # print a specific day
stickies days                  # list all journal dates
stickies search "todo"         # grep across stickies + journal
```

Data lives in `~/.local/share/stickies/` by default (overridable via `$STICKIES_DIR` or `--data`). The `$EDITOR` env var is honored when editing notes externally (`o` key).

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
