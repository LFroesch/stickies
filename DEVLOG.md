# DEVLOG

## DevLog

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
