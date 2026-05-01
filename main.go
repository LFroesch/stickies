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

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	dataDir := flag.String("data", "", "Override data directory")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "stickies — quick notes + daily journal TUI\n\n")
		fmt.Fprintf(os.Stderr, "Usage: stickies [flags]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("stickies " + version)
		return
	}

	dir := *dataDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dir = filepath.Join(home, ".local", "share", "stickies")
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
