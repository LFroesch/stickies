# DEVLOG

## DevLog

### 2026-05-04 — `stk` alias + `new`/`log` subcommands

- **`stk` alias** (`Makefile`, `install.sh`): both install paths now drop a `stk` symlink next to the `stickies` binary in `$INSTALL_DIR`. Lets you type `stk ls`, `stk new "..."`, etc.
- **`stickies new <text>`** (`cli.go`, `main.go`): create a sticky directly from the CLI. Reads from arg or stdin (`echo "..." | stk new`). Prints the new sticky's short ID + first line.
- **`stickies log <text>`** (`cli.go`, `main.go`): append a timestamped bullet (`- HH:MM <text>`) to today's journal entry, creating it if missing. Stdin supported the same way.
- **Help text updated** to surface the new subcommands; `--version`/`--help` already worked via Go's flag package.

### 2026-05-04 — v1 fix: cancel-safe new stickies, editor keybind discoverability

- **Canceling a new sticky no longer leaves a phantom note behind** (`model.go`, `update.go`): starting a new sticky still drops you straight into inline editing, but hitting `esc` now removes the unsaved placeholder from memory instead of leaving behind a note that was never written to disk.
- **Edit-mode keybinds are surfaced in the UI and docs** (`view.go`, `README.md`): the footer/help now call out existing textarea bindings like `ctrl+d`, `home/end`, `ctrl+a/e`, `alt+b/f`, and word delete motions so inline editing feels closer to a real text editor.
- **Quoted editor commands now work** (`helpers.go`, `helpers_test.go`, `update.go`, `README.md`): external editor launch no longer breaks on values like `EDITOR='code --wait'` or escaped spaces in editor arguments.

### 2026-05-01 — v1 polish: pipeable CLI, word-wrap, tab nav

- **Pipe-friendly subcommands** (`cli.go`, `main.go`): added `stickies ls`, `cat <id|title>`, `today`, `day <YYYY-MM-DD>`, `days`, `search <query>`. Stickies now have a stable short ID (8-hex prefix) and `cat` accepts id, id-prefix, or case-insensitive title substring. Journal dates were already filename-friendly. Lets you `stickies today | grep TODO`, `stickies cat shopping | wc -l`, etc.
- **`$STICKIES_DIR` env override**: data dir resolution now honors `$STICKIES_DIR` ahead of the default `~/.local/share/stickies/`. `--data` still wins over the env var.
- **Word-aware wrapping** (`view.go`): `wrapText` previously hard-broke at any rune boundary, splitting mid-word in the detail/journal panes. Now it wraps on word boundaries and only character-breaks for words longer than the panel width. Leading whitespace of a line is preserved.
- **Tab navigation** (`update.go`): `tab` / `shift+tab` cycle 1→2→3→1 across stickies/journal/search. README claimed this worked; it didn't. Now it does.
- **Help panel surfaces settings** (`view.go`): the help overlay now shows the active config dir plus `$EDITOR` / `$STICKIES_DIR` overrides, so users can find their data without `cat`-ing the source.
- **CLAUDE.md added**: project conventions + file roles for tui-suite parity.
- **README**: renamed "Quick Install" → "Install" so it matches the suite-wide audit, added Screenshot placeholder, version line, and `## Environment` section enumerating env vars.

### Pre-v1
Initial Bubbletea implementation: stickies + journal pages, frontmatter markdown persistence, search across both, in-app textarea editor, `$EDITOR` integration via `tea.ExecProcess`, pin/tag/yank/delete flows.
