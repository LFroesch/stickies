# Stickies

Quick notes and a daily journal in one TUI. Pinnable stickies on the left, dated journal entries on the right, fast keyboard editing for both. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Quick Install

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
stickies                # launch the TUI
stickies --version
stickies --data ./mydir # override data directory
```

Data lives in `~/.local/share/stickies/` by default. The `$EDITOR` env var is honored when editing notes externally (`e` key).

## Keybinds

| Key            | Action                              |
|----------------|-------------------------------------|
| `q`            | Quit                                |
| `?`            | Toggle help                         |
| `tab`          | Switch between stickies and journal |
| `j`/`k`, `↑/↓` | Move selection                      |
| `g` / `G`      | Jump to top / bottom                |
| `n`            | New sticky                          |
| `t`            | New tagged sticky                   |
| `enter` / `E`  | Edit selected (in-app)              |
| `e`            | Edit selected in `$EDITOR`          |
| `o`            | Open in default app                 |
| `y`            | Yank/copy contents                  |
| `p`            | Toggle pin                          |
| `D`            | Delete (with confirm)               |
| `/`            | Search/filter                       |
| `esc`          | Cancel mode / close overlay         |

## Configuration

| Env var             | Effect                                    |
|---------------------|-------------------------------------------|
| `EDITOR`            | External editor for `e` keybind           |
| `--data <dir>`      | Override data directory at launch         |

## License

GPL-3.0 — see [LICENSE](LICENSE).
