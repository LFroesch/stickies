package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "v1.0.0"

func main() {
	// Subcommand dispatch (positional verb before flags) for piping.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ls", "list", "cat", "show", "get", "today", "day", "days", "search", "grep", "new", "n", "log", "l":
			dir, err := resolveDataDir(envDataDir())
			if err != nil {
				log.Fatal(err)
			}
			if err := ensureDirs(dir); err != nil {
				log.Fatal(err)
			}
			os.Exit(runCLI(os.Args[1], os.Args[2:], dir, os.Stdout))
		}
	}

	showVersion := flag.Bool("version", false, "Print version and exit")
	dataDir := flag.String("data", "", "Override data directory (default: $STICKIES_DIR or ~/.local/share/stickies)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "stickies — quick notes + daily journal TUI\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  stickies [flags]              launch TUI\n")
		fmt.Fprintf(os.Stderr, "  stickies ls                   list stickies (id\\ttitle)\n")
		fmt.Fprintf(os.Stderr, "  stickies cat <id|title>       print sticky body to stdout\n")
		fmt.Fprintf(os.Stderr, "  stickies today                print today's journal entry\n")
		fmt.Fprintf(os.Stderr, "  stickies day <YYYY-MM-DD>     print specified day's entry\n")
		fmt.Fprintf(os.Stderr, "  stickies days                 list journal dates\n")
		fmt.Fprintf(os.Stderr, "  stickies search <query>       grep across stickies+journal\n")
		fmt.Fprintf(os.Stderr, "  stickies new <text>           new sticky from arg or stdin\n")
		fmt.Fprintf(os.Stderr, "  stickies log <text>           append timestamped line to today's journal\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, "stickies "+version)
		return
	}

	override := *dataDir
	if override == "" {
		override = envDataDir()
	}
	dir, err := resolveDataDir(override)
	if err != nil {
		log.Fatal(err)
	}
	if err := ensureDirs(dir); err != nil {
		log.Fatal(err)
	}

	stickies, err := loadStickies(dir)
	if err != nil {
		log.Fatal(err)
	}
	journal, err := loadJournal(dir)
	if err != nil {
		log.Fatal(err)
	}

	body := textarea.New()
	body.CharLimit = 0
	body.ShowLineNumbers = false
	body.Prompt = ""

	meta := textinput.New()
	meta.Placeholder = "tags (comma-separated)"
	meta.CharLimit = 200

	search := textinput.New()
	search.Placeholder = "search…"
	search.CharLimit = 100

	m := model{
		dataDir:       dir,
		width:         100,
		height:        24,
		page:          pageStickies,
		mode:          modeNormal,
		focus:         focusLeft,
		stickies:      stickies,
		journal:       journal,
		bodyArea:      body,
		metaInput:     meta,
		searchInput:   search,
		editingSticky: -1,
		deleteIndex:   -1,
	}
	sortStickies(m.stickies)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func envDataDir() string { return os.Getenv("STICKIES_DIR") }

func resolveDataDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "stickies"), nil
}
