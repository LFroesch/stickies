package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

type appPage int

const (
	pageStickies appPage = iota
	pageJournal
	pageSearch
)

type appMode int

const (
	modeNormal appMode = iota
	modeBodyEdit
	modeMetaEdit
	modeDeleteConfirm
	modeSearch
	modeHelp
)

type focusPanel int

const (
	focusLeft focusPanel = iota
	focusRight
)

type metaField int

const (
	metaFieldTags metaField = iota
)

type Sticky struct {
	ID      string
	Body    string
	Pinned  bool
	Tags    []string
	Created time.Time
	Updated time.Time
}

type JournalEntry struct {
	Date    string
	Body    string
	Updated time.Time
}

type model struct {
	dataDir string
	width   int
	height  int

	page  appPage
	mode  appMode
	focus focusPanel

	stickies      []Sticky
	stickyCursor  int
	stickyScroll  int

	journal       []JournalEntry
	journalCursor int
	journalScroll int

	bodyArea       textarea.Model
	editingSticky  int // index into stickies, -1 if editing journal
	editingDate    string

	metaInput textinput.Model
	metaField metaField

	searchInput   textinput.Model
	searchQuery   string
	searchResults []searchHit

	deleteIndex   int
	deleteIsJournal bool

	helpScroll int

	statusMsg    string
	statusExpiry time.Time
}

type searchHit struct {
	isJournal bool
	stickyIdx int
	date      string
	preview   string
}
