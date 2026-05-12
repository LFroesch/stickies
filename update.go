package main

import (
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("stickies")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeBodyArea()
		return m, nil

	case statusExpiryMsg:
		if !m.statusExpiry.IsZero() && time.Now().After(m.statusExpiry) {
			m.statusMsg = ""
			m.statusExpiry = time.Time{}
		}
		return m, nil

	case editorClosedMsg:
		// Reload after $EDITOR session.
		if msg.err != nil {
			return m, m.flash("editor: " + msg.err.Error())
		}
		if msg.isJournal {
			j, err := loadJournal(m.dataDir)
			if err == nil {
				m.journal = j
				if i := findJournal(m.journal, msg.date); i >= 0 {
					m.journalCursor = i
				}
			}
		} else {
			s, err := loadStickies(m.dataDir)
			if err == nil {
				m.stickies = s
				sortStickies(m.stickies)
			}
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	switch m.mode {
	case modeBodyEdit:
		var cmd tea.Cmd
		m.bodyArea, cmd = m.bodyArea.Update(msg)
		return m, cmd
	case modeMetaEdit:
		var cmd tea.Cmd
		m.metaInput, cmd = m.metaInput.Update(msg)
		return m, cmd
	case modeSearch:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.searchQuery = m.searchInput.Value()
		m.searchResults = searchAll(m.stickies, m.journal, m.searchQuery)
		return m, cmd
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.mode {
	case modeBodyEdit:
		return m.handleBodyEditKey(msg)
	case modeMetaEdit:
		return m.handleMetaEditKey(msg)
	case modeDeleteConfirm:
		return m.handleDeleteKey(msg)
	case modeSearch:
		return m.handleSearchKey(msg)
	case modeHelp:
		return m.handleHelpKey(msg)
	}

	return m.handleNormalKey(msg)
}

// --- Normal mode ---

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "?":
		m.mode = modeHelp
		m.helpScroll = 0
		return m, nil
	case "1":
		m.page = pageStickies
		return m, nil
	case "2":
		m.page = pageJournal
		m.ensureJournalCursor()
		return m, nil
	case "3":
		m.page = pageSearch
		m.mode = modeSearch
		m.searchInput.Focus()
		return m, nil
	case "tab":
		m.page = (m.page + 1) % 3
		if m.page == pageJournal {
			m.ensureJournalCursor()
		} else if m.page == pageSearch {
			m.mode = modeSearch
			m.searchInput.Focus()
		}
		return m, nil
	case "shift+tab":
		m.page = (m.page + 2) % 3
		if m.page == pageJournal {
			m.ensureJournalCursor()
		} else if m.page == pageSearch {
			m.mode = modeSearch
			m.searchInput.Focus()
		}
		return m, nil
	case "/":
		m.page = pageSearch
		m.mode = modeSearch
		m.searchInput.Focus()
		return m, nil
	case "n":
		return m.startNewSticky()
	case "t":
		return m.openTodayJournal()
	}

	switch m.page {
	case pageStickies:
		return m.handleStickiesKey(msg)
	case pageJournal:
		return m.handleJournalKey(msg)
	case pageSearch:
		return m.handleSearchResultsKey(msg)
	}
	return m, nil
}

func (m model) handleStickiesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if len(m.stickies) > 0 {
			m.stickyCursor = clamp(m.stickyCursor+1, 0, len(m.stickies)-1)
		}
	case "k", "up":
		if len(m.stickies) > 0 {
			m.stickyCursor = clamp(m.stickyCursor-1, 0, len(m.stickies)-1)
		}
	case "g":
		m.stickyCursor = 0
	case "G":
		if len(m.stickies) > 0 {
			m.stickyCursor = len(m.stickies) - 1
		}
	case "ctrl+d":
		m.stickyCursor = clamp(m.stickyCursor+m.listPageStep(), 0, max(0, len(m.stickies)-1))
	case "ctrl+u":
		m.stickyCursor = clamp(m.stickyCursor-m.listPageStep(), 0, max(0, len(m.stickies)-1))
	case "p":
		if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
			m.stickies[i].Pinned = !m.stickies[i].Pinned
			m.stickies[i].Updated = time.Now()
			id := m.stickies[i].ID
			if err := saveSticky(m.dataDir, m.stickies[i]); err != nil {
				return m, m.flash("save failed: " + err.Error())
			}
			sortStickies(m.stickies)
			for k, s := range m.stickies {
				if s.ID == id {
					m.stickyCursor = k
					break
				}
			}
			return m, m.flash("toggled pin")
		}
	case "E", "enter":
		return m.editStickyBody()
	case "e":
		return m.editStickyMeta()
	case "o", "O":
		return m.openStickyInEditor()
	case "y":
		if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
			if err := clipboard.WriteAll(m.stickies[i].Body); err != nil {
				return m, m.flash("clipboard: " + err.Error())
			}
			return m, m.flash("copied")
		}
	case "D":
		if len(m.stickies) > 0 {
			m.mode = modeDeleteConfirm
			m.deleteIndex = m.stickyCursor
			m.deleteIsJournal = false
		}
	}
	return m, nil
}

func (m model) handleJournalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if len(m.journal) > 0 {
			m.journalCursor = clamp(m.journalCursor+1, 0, len(m.journal)-1)
		}
	case "k", "up":
		if len(m.journal) > 0 {
			m.journalCursor = clamp(m.journalCursor-1, 0, len(m.journal)-1)
		}
	case "g":
		m.journalCursor = 0
	case "G":
		if len(m.journal) > 0 {
			m.journalCursor = len(m.journal) - 1
		}
	case "ctrl+d":
		m.journalCursor = clamp(m.journalCursor+m.listPageStep(), 0, max(0, len(m.journal)-1))
	case "ctrl+u":
		m.journalCursor = clamp(m.journalCursor-m.listPageStep(), 0, max(0, len(m.journal)-1))
	case "E", "enter":
		return m.editJournalBody()
	case "o", "O":
		return m.openJournalInEditor()
	case "y":
		if i := m.journalCursor; i >= 0 && i < len(m.journal) {
			if err := clipboard.WriteAll(m.journal[i].Body); err != nil {
				return m, m.flash("clipboard: " + err.Error())
			}
			return m, m.flash("copied")
		}
	case "D":
		if len(m.journal) > 0 {
			m.mode = modeDeleteConfirm
			m.deleteIndex = m.journalCursor
			m.deleteIsJournal = true
		}
	}
	return m, nil
}

func (m model) handleSearchResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.page = pageStickies
		m.mode = modeNormal
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.searchQuery = ""
		m.searchResults = nil
		return m, nil
	case "/":
		m.mode = modeSearch
		m.searchInput.Focus()
	}
	return m, nil
}

// --- Body edit mode ---

func (m model) handleBodyEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.discardPendingNewSticky()
		m.bodyArea.Blur()
		m.mode = modeNormal
		m.editingSticky = -1
		m.editingDate = ""
		m.creatingSticky = false
		return m, nil
	case "ctrl+s":
		return m.commitBodyEdit()
	}
	var cmd tea.Cmd
	m.bodyArea, cmd = m.bodyArea.Update(msg)
	return m, cmd
}

func (m model) commitBodyEdit() (tea.Model, tea.Cmd) {
	body := m.bodyArea.Value()
	now := time.Now()
	if m.editingDate != "" {
		e := JournalEntry{Date: m.editingDate, Body: body, Updated: now}
		if err := saveJournal(m.dataDir, e); err != nil {
			return m, m.flash("save failed: " + err.Error())
		}
		m.journal = upsertJournal(m.journal, e)
		if i := findJournal(m.journal, m.editingDate); i >= 0 {
			m.journalCursor = i
		}
	} else if i := m.editingSticky; i >= 0 && i < len(m.stickies) {
		m.stickies[i].Body = body
		m.stickies[i].Updated = now
		if m.stickies[i].Created.IsZero() {
			m.stickies[i].Created = now
		}
		if err := saveSticky(m.dataDir, m.stickies[i]); err != nil {
			return m, m.flash("save failed: " + err.Error())
		}
		id := m.stickies[i].ID
		sortStickies(m.stickies)
		for k, s := range m.stickies {
			if s.ID == id {
				m.stickyCursor = k
				break
			}
		}
	}
	m.bodyArea.Blur()
	m.mode = modeNormal
	m.editingSticky = -1
	m.editingDate = ""
	m.creatingSticky = false
	return m, m.flash("saved")
}

// --- Meta edit mode ---

func (m model) handleMetaEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.metaInput.Blur()
		m.mode = modeNormal
		return m, nil
	case "enter":
		if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
			m.stickies[i].Tags = parseTags(m.metaInput.Value())
			m.stickies[i].Updated = time.Now()
			if err := saveSticky(m.dataDir, m.stickies[i]); err != nil {
				return m, m.flash("save failed: " + err.Error())
			}
		}
		m.metaInput.Blur()
		m.mode = modeNormal
		return m, m.flash("tags saved")
	}
	var cmd tea.Cmd
	m.metaInput, cmd = m.metaInput.Update(msg)
	return m, cmd
}

// --- Delete confirm ---

func (m model) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.deleteIsJournal {
			if i := m.deleteIndex; i >= 0 && i < len(m.journal) {
				if err := deleteJournal(m.dataDir, m.journal[i].Date); err != nil {
					return m, m.flash("delete failed: " + err.Error())
				}
				m.journal = append(m.journal[:i], m.journal[i+1:]...)
				m.journalCursor = clamp(m.journalCursor, 0, max(0, len(m.journal)-1))
			}
		} else {
			if i := m.deleteIndex; i >= 0 && i < len(m.stickies) {
				if err := deleteSticky(m.dataDir, m.stickies[i].ID); err != nil {
					return m, m.flash("delete failed: " + err.Error())
				}
				m.stickies = append(m.stickies[:i], m.stickies[i+1:]...)
				m.stickyCursor = clamp(m.stickyCursor, 0, max(0, len(m.stickies)-1))
			}
		}
		m.mode = modeNormal
		m.deleteIndex = -1
		return m, m.flash("deleted")
	case "esc", "n", "N":
		m.mode = modeNormal
		m.deleteIndex = -1
	}
	return m, nil
}

// --- Search input mode ---

func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.searchInput.Blur()
		return m, nil
	case "enter":
		m.mode = modeNormal
		m.searchInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = m.searchInput.Value()
	m.searchResults = searchAll(m.stickies, m.journal, m.searchQuery)
	return m, cmd
}

// --- Help ---

func (m model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.mode = modeNormal
	case "j", "down":
		m.helpScroll++
	case "k", "up":
		if m.helpScroll > 0 {
			m.helpScroll--
		}
	case "g":
		m.helpScroll = 0
	}
	return m, nil
}

// --- Action helpers ---

func (m model) listPageStep() int {
	step := m.bodyHeight()/2 - 1
	if step < 1 {
		return 1
	}
	return step
}

// resizeBodyArea recomputes the textarea dimensions to fit the right panel
// inside the current page layout (stickies vs. journal).
func (m *model) resizeBodyArea() {
	bodyH := m.height - 2 // header + footer
	if bodyH < 6 {
		bodyH = 6
	}
	innerH := bodyH - 2 // panel borders

	var leftW int
	switch m.page {
	case pageJournal:
		leftW = clamp(m.width*22/100, 18, 28)
	default:
		leftW = clamp(m.width*30/100, 28, 44)
	}
	rightW := m.width - leftW
	innerW := rightW - 4
	if innerW < 10 {
		innerW = 10
	}

	// reserve: 1 line for date/sticky header inside the panel-content area
	taH := innerH - 1 - 1 // panel-label + content header
	if m.page == pageJournal && m.editingDate != "" {
		yLines := m.yesterdayPreviewLines(innerW, 5)
		if len(yLines) > 0 {
			taH -= len(yLines) + 1 // separator + lines
		}
	}
	if taH < 3 {
		taH = 3
	}
	m.bodyArea.SetWidth(innerW)
	m.bodyArea.SetHeight(taH)
}

func (m *model) discardPendingNewSticky() {
	if !m.creatingSticky {
		return
	}
	if i := m.editingSticky; i >= 0 && i < len(m.stickies) {
		m.stickies = append(m.stickies[:i], m.stickies[i+1:]...)
		m.stickyCursor = clamp(m.stickyCursor, 0, max(0, len(m.stickies)-1))
	}
}

func (m model) startNewSticky() (tea.Model, tea.Cmd) {
	now := time.Now()
	s := Sticky{
		ID:      newID(),
		Created: now,
		Updated: now,
	}
	m.stickies = append([]Sticky{s}, m.stickies...)
	sortStickies(m.stickies)
	for i, x := range m.stickies {
		if x.ID == s.ID {
			m.stickyCursor = i
			break
		}
	}
	m.page = pageStickies
	m.editingSticky = m.stickyCursor
	m.editingDate = ""
	m.bodyArea.SetValue("")
	m.resizeBodyArea()
	m.bodyArea.Focus()
	m.mode = modeBodyEdit
	m.creatingSticky = true
	return m, nil
}

func (m model) editStickyBody() (tea.Model, tea.Cmd) {
	if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
		m.editingSticky = i
		m.editingDate = ""
		m.bodyArea.SetValue(m.stickies[i].Body)
		m.resizeBodyArea()
		m.bodyArea.Focus()
		m.mode = modeBodyEdit
		m.creatingSticky = false
	}
	return m, nil
}

func (m model) editStickyMeta() (tea.Model, tea.Cmd) {
	if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
		m.metaInput.SetValue(strings.Join(m.stickies[i].Tags, ", "))
		m.metaInput.Focus()
		m.mode = modeMetaEdit
	}
	return m, nil
}

func (m model) openStickyInEditor() (tea.Model, tea.Cmd) {
	if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
		path := stickyPath(m.dataDir, m.stickies[i].ID)
		return m, runEditor(path, false, m.stickies[i].ID)
	}
	return m, nil
}

func (m model) editJournalBody() (tea.Model, tea.Cmd) {
	if i := m.journalCursor; i >= 0 && i < len(m.journal) {
		m.editingDate = m.journal[i].Date
		m.editingSticky = -1
		m.creatingSticky = false
		m.bodyArea.SetValue(m.journal[i].Body)
		m.resizeBodyArea()
		m.bodyArea.Focus()
		m.mode = modeBodyEdit
	}
	return m, nil
}

func (m model) openTodayJournal() (tea.Model, tea.Cmd) {
	m.page = pageJournal
	date := todayDate()
	if i := findJournal(m.journal, date); i >= 0 {
		m.journalCursor = i
		m.editingDate = date
		m.editingSticky = -1
		m.creatingSticky = false
		m.bodyArea.SetValue(m.journal[i].Body)
	} else {
		e := JournalEntry{Date: date, Updated: time.Now()}
		m.journal = upsertJournal(m.journal, e)
		m.journalCursor = findJournal(m.journal, date)
		m.editingDate = date
		m.editingSticky = -1
		m.creatingSticky = false
		m.bodyArea.SetValue("")
	}
	m.resizeBodyArea()
	m.bodyArea.Focus()
	m.mode = modeBodyEdit
	return m, nil
}

func (m model) openJournalInEditor() (tea.Model, tea.Cmd) {
	if i := m.journalCursor; i >= 0 && i < len(m.journal) {
		date := m.journal[i].Date
		// Ensure file exists for $EDITOR.
		if _, err := loadJournal(m.dataDir); err == nil {
			if findJournal(m.journal, date) < 0 || m.journal[i].Updated.IsZero() {
				_ = saveJournal(m.dataDir, m.journal[i])
			}
		}
		path := journalPath(m.dataDir, date)
		return m, runEditor(path, true, date)
	}
	return m, nil
}

func (m model) ensureJournalCursor() {
	today := todayDate()
	yesterday := yesterdayDate()
	if i := findJournal(m.journal, today); i >= 0 {
		m.journalCursor = i
		return
	}
	if i := findJournal(m.journal, yesterday); i >= 0 {
		m.journalCursor = i
		return
	}
	m.journalCursor = 0
}

// --- Status flash ---

type statusExpiryMsg struct{}

func (m *model) flash(msg string) tea.Cmd {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(3 * time.Second)
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return statusExpiryMsg{} })
}

// --- Editor session ---

type editorClosedMsg struct {
	err       error
	isJournal bool
	date      string
}

func runEditor(path string, isJournal bool, date string) tea.Cmd {
	editor := resolveEditor()
	parts, err := splitCommandLine(editor)
	if err != nil {
		return func() tea.Msg {
			return editorClosedMsg{err: err, isJournal: isJournal, date: date}
		}
	}
	parts = append(parts, path)
	cmd := exec.Command(parts[0], parts[1:]...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorClosedMsg{err: err, isJournal: isJournal, date: date}
	})
}
