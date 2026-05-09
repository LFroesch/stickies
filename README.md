# stickies

Quick notes and a daily journal in one terminal app. `stickies` has a full-screen TUI for browsing and editing, plus a small CLI for piping notes in and reading them back out.

## Install

Supported platforms: Linux and macOS. On Windows, use WSL.

Recommended:

```bash
curl -fsSL https://raw.githubusercontent.com/LFroesch/stickies/main/install.sh | bash
```

Other options:

```bash
go install github.com/LFroesch/stickies@latest
make install
```

Commands:

```bash
stickies
stk
stk --version
stk --help
```

`stk` is installed as a short alias for the same binary.

## TUI

- Stickies on the left
- Journal on the right
- Search across both
- Fast in-app editing or open in `$VISUAL` / `$EDITOR`
- Pinned notes stay easy to reach while journal entries remain date-based

Default data directory:

```text
~/.local/share/stickies/
```

Override it with `$STICKIES_DIR` or `--data`.

## CLI

```bash
stk ls
stk cat <id|title>
stk new "buy milk"
stk log "finished cli wiring"
stk today
stk day 2026-04-01
stk days
stk search "todo"

echo "multi-line note" | stk new
git log --oneline -5 | stk log
```

The CLI is meant to stay pipe-friendly, so `new` and `log` both accept stdin.

## Controls

| Key | Action |
|-----|--------|
| `1`, `2`, `3` | Stickies, Journal, Search |
| `tab` | Change page |
| `j/k` | Move |
| `n` | New sticky |
| `t` | Jump to today's journal |
| `enter`, `E` | Edit body |
| `e` | Edit tags |
| `o` | Open in external editor |
| `p` | Toggle pin |
| `D` | Delete |
| `/` | Search |
| `ctrl+s` | Save in edit mode |
| `?` | Help |
| `q` | Quit |

## Environment

| Var or Flag | Purpose |
|-------------|---------|
| `VISUAL` | preferred external editor |
| `EDITOR` | fallback external editor |
| `STICKIES_DIR` | override data directory |
| `--data <dir>` | override data directory |

## License

[GPL-3.0](LICENSE)
