package main

import (
	"fmt"
	"strings"

	"github.com/LFroesch/tui-suite/suitechrome"
	"github.com/charmbracelet/lipgloss"
)

const (
	minTerminalW = 60
	minTerminalH = 16
)

var (
	bgColor       = lipgloss.Color("235")
	headerFg      = lipgloss.Color("214")
	statusKeyFg   = lipgloss.Color("214")
	statusValueFg = lipgloss.Color("255")
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var body string
	if m.mode == modeHelp {
		body = m.renderHelpPanel()
	} else {
		body = m.renderBody()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// --- Header ---

func (m model) renderHeader() string {
	title := suitechrome.RenderTitle("stickies", version)
	tabs := suitechrome.RenderTabs([]suitechrome.Tab{
		{Label: "1 stickies", Active: m.page == pageStickies},
		{Label: "2 journal", Active: m.page == pageJournal},
		{Label: "3 search", Active: m.page == pageSearch},
	})

	var count string
	switch m.page {
	case pageStickies:
		count = suitechrome.Dim(fmt.Sprintf("%d notes", len(m.stickies)))
	case pageJournal:
		count = suitechrome.Dim(fmt.Sprintf("%d days", len(m.journal)))
	case pageSearch:
		count = suitechrome.Dim(fmt.Sprintf("%d hits", len(m.searchResults)))
	}

	return suitechrome.JoinHeader(m.width, title+"  "+tabs, count)
}

// --- Footer ---

func (m model) renderFooter() string {
	hint := m.footerHint()
	status := ""
	if m.statusMsg != "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true).Inline(true).Render(m.statusMsg)
	}
	return suitechrome.JoinLine(m.width, hint, status)
}

func (m model) footerHint() string {
	pair := func(k, d string) suitechrome.Action { return suitechrome.Action{Key: k, Label: d} }
	join := func(parts ...suitechrome.Action) string { return suitechrome.RenderActions(parts) }

	switch m.mode {
	case modeBodyEdit:
		return join(pair("ctrl+s", "save"), pair("ctrl+d", "del row"), pair("home/end", "bol/eol"), pair("esc", "cancel"))
	case modeMetaEdit:
		return join(pair("enter", "save"), pair("esc", "cancel"))
	case modeDeleteConfirm:
		warn := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Inline(true)
		return warn.Render("delete? ") + join(pair("y", "yes"), pair("n/esc", "no"))
	case modeSearch:
		return join(pair("enter", "lock"), pair("esc", "exit"))
	}
	switch m.page {
	case pageStickies:
		return join(pair("n", "new"), pair("E", "edit"), pair("e", "tags"),
			pair("p", "pin"), pair("o", "$EDITOR"), pair("y", "yank"),
			pair("D", "del"), pair("/", "search"), pair("?", "help"), pair("q", "quit"))
	case pageJournal:
		return join(pair("t", "today"), pair("E", "edit"), pair("o", "$EDITOR"),
			pair("y", "yank"), pair("D", "del"), pair("/", "search"),
			pair("?", "help"), pair("q", "quit"))
	case pageSearch:
		return join(pair("/", "refine"), pair("esc", "back"), pair("?", "help"), pair("q", "quit"))
	}
	return ""
}

// --- Body dispatch ---

func (m model) renderBody() string {
	panelH := m.bodyHeight()
	switch m.page {
	case pageStickies:
		return m.renderStickies(panelH)
	case pageJournal:
		return m.renderJournal(panelH)
	case pageSearch:
		return m.renderSearch(panelH)
	}
	return ""
}

// bodyHeight = total rows the body occupies (including panel borders).
func (m model) bodyHeight() int {
	h := m.height - 2 // header + footer
	if h < 6 {
		h = 6
	}
	return h
}

// --- Stickies page ---

func (m model) renderStickies(h int) string {
	leftW := clamp(m.width*30/100, 28, 44)
	rightW := m.width - leftW

	leftActive := m.focus == focusLeft && m.mode == modeNormal
	rightActive := m.mode == modeBodyEdit && m.editingDate == "" && m.editingSticky >= 0

	leftBox := panelBox(leftActive, leftW, h, "stickies", m.renderStickyList(leftW-4, h-2))
	rightBox := panelBox(rightActive, rightW, h, "detail", m.renderStickyDetail(rightW-4, h-2))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (m model) renderStickyList(w, h int) string {
	if len(m.stickies) == 0 {
		return dimTextStyle.Render("no stickies\n\npress n to add")
	}
	max := h - 1 // reserve panel header line
	if max < 1 {
		max = 1
	}

	start := 0
	if m.stickyCursor >= max {
		start = m.stickyCursor - max + 1
	}
	end := start + max
	if end > len(m.stickies) {
		end = len(m.stickies)
	}

	var lines []string
	for i := start; i < end; i++ {
		s := m.stickies[i]
		marker := "  "
		if i == m.stickyCursor {
			marker = "▸ "
		}
		pin := "  "
		if s.Pinned {
			pin = pinnedMarkStyle.Render("★ ")
		}
		preview := firstLine(s.Body)
		if preview == "" {
			preview = dimTextStyle.Render("(empty)")
		}
		line := marker + pin + truncate(preview, w-4)
		if i == m.stickyCursor {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Width(w).
				Render(line)
		}
		lines = append(lines, line)
	}
	if start > 0 {
		lines[0] = dimTextStyle.Render(fmt.Sprintf("▲ %d more", start))
	}
	if end < len(m.stickies) {
		if len(lines) >= max {
			lines[len(lines)-1] = dimTextStyle.Render(fmt.Sprintf("▼ %d more", len(m.stickies)-end))
		} else {
			lines = append(lines, dimTextStyle.Render(fmt.Sprintf("▼ %d more", len(m.stickies)-end)))
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) renderStickyDetail(w, h int) string {
	if m.mode == modeBodyEdit && m.editingSticky >= 0 && m.editingDate == "" {
		return m.bodyArea.View()
	}
	if i := m.stickyCursor; i >= 0 && i < len(m.stickies) {
		s := m.stickies[i]
		var b strings.Builder
		header := ""
		if s.Pinned {
			header = pinnedMarkStyle.Render("★ pinned") + "  "
		}
		header += dimTextStyle.Render(s.Updated.Format("2006-01-02 15:04"))
		b.WriteString(header)
		b.WriteString("\n")
		if len(s.Tags) > 0 {
			var chips []string
			for _, t := range s.Tags {
				chips = append(chips, tagChipStyle.Render("#"+t))
			}
			b.WriteString(strings.Join(chips, " "))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		body := s.Body
		if body == "" {
			body = dimTextStyle.Render("(empty — press E to edit)")
		}
		b.WriteString(wrapText(body, w))

		if m.mode == modeMetaEdit {
			b.WriteString("\n\n")
			b.WriteString(keyStyle.Render("tags: "))
			b.WriteString(m.metaInput.View())
		}
		return clipLines(b.String(), h)
	}
	return dimTextStyle.Render("press n to add a sticky")
}

// --- Journal page ---

func (m model) renderJournal(h int) string {
	leftW := clamp(m.width*22/100, 18, 28)
	rightW := m.width - leftW

	leftActive := m.focus == focusLeft && m.mode == modeNormal
	rightActive := m.mode == modeBodyEdit && m.editingDate != ""

	leftBox := panelBox(leftActive, leftW, h, "days", m.renderJournalList(leftW-4, h-2))
	rightBox := panelBox(rightActive, rightW, h, "entry", m.renderJournalDetail(rightW-4, h-2))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (m model) renderJournalList(w, h int) string {
	if len(m.journal) == 0 {
		return dimTextStyle.Render("no entries\n\npress t for today")
	}
	today := todayDate()
	max := h - 1
	if max < 1 {
		max = 1
	}
	start := 0
	if m.journalCursor >= max {
		start = m.journalCursor - max + 1
	}
	end := start + max
	if end > len(m.journal) {
		end = len(m.journal)
	}

	var lines []string
	for i := start; i < end; i++ {
		e := m.journal[i]
		marker := "  "
		if i == m.journalCursor {
			marker = "▸ "
		}
		var label string
		if e.Date == today {
			label = todayMarkStyle.Render(e.Date + " *")
		} else {
			label = dateHeaderStyle.Render(e.Date)
		}
		line := marker + label
		if i == m.journalCursor {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Width(w).
				Render(line)
		}
		lines = append(lines, line)
	}
	if start > 0 {
		lines[0] = dimTextStyle.Render(fmt.Sprintf("▲ %d more", start))
	}
	if end < len(m.journal) {
		if len(lines) >= max {
			lines[len(lines)-1] = dimTextStyle.Render(fmt.Sprintf("▼ %d more", len(m.journal)-end))
		} else {
			lines = append(lines, dimTextStyle.Render(fmt.Sprintf("▼ %d more", len(m.journal)-end)))
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) renderJournalDetail(w, h int) string {
	if m.mode == modeBodyEdit && m.editingDate != "" {
		header := dateHeaderStyle.Render(m.editingDate)
		yLines := m.yesterdayPreviewLines(w, 5)
		yBlock := ""
		if len(yLines) > 0 {
			yBlock = "\n" + dimTextStyle.Render("─ yesterday "+strings.Repeat("─", clamp(w-12, 0, 40))) + "\n" +
				dimTextStyle.Render(strings.Join(yLines, "\n"))
		}
		// editor view; bodyArea sized in update.go to fit panel
		out := header + "\n" + m.bodyArea.View() + yBlock
		return clipLines(out, h)
	}
	if i := m.journalCursor; i >= 0 && i < len(m.journal) {
		e := m.journal[i]
		header := dateHeaderStyle.Render(e.Date)
		if e.Date == todayDate() {
			header += "  " + todayMarkStyle.Render("today")
		}
		body := e.Body
		if body == "" {
			body = dimTextStyle.Render("(empty — press E to edit, t for today)")
		}
		out := header + "\n\n" + wrapText(body, w)
		return clipLines(out, h)
	}
	return dimTextStyle.Render("press t to start today's entry")
}

func (m model) yesterdayPreviewLines(w, maxLines int) []string {
	y := yesterdayDate()
	i := findJournal(m.journal, y)
	if i < 0 {
		return nil
	}
	body := strings.TrimSpace(m.journal[i].Body)
	if body == "" {
		return nil
	}
	wrapped := strings.Split(wrapText(body, w), "\n")
	if len(wrapped) > maxLines {
		wrapped = append(wrapped[:maxLines-1], "…")
	}
	return wrapped
}

// --- Search page ---

func (m model) renderSearch(h int) string {
	body := m.renderSearchInner(m.width-4, h-2)
	return panelBox(m.mode == modeSearch, m.width, h, "search", body)
}

func (m model) renderSearchInner(w, h int) string {
	q := m.searchInput.View()
	if m.mode != modeSearch {
		q = dimTextStyle.Render("query: ") + m.searchQuery
	}
	header := keyStyle.Render("/ ") + q
	if len(m.searchResults) == 0 {
		body := dimTextStyle.Render("no matches")
		if m.searchQuery == "" {
			body = dimTextStyle.Render("type to search across stickies + journal")
		}
		return clipLines(header+"\n\n"+body, h)
	}
	max := h - 2
	if max < 1 {
		max = 1
	}
	results := m.searchResults
	if len(results) > max {
		results = results[:max]
	}
	var lines []string
	for _, hit := range results {
		var prefix string
		if hit.isJournal {
			prefix = dateHeaderStyle.Render("[" + hit.date + "]")
		} else if hit.stickyIdx >= 0 && hit.stickyIdx < len(m.stickies) {
			s := m.stickies[hit.stickyIdx]
			tag := "sticky"
			if s.Pinned {
				tag = "★"
			}
			prefix = tagChipStyle.Render("[" + tag + "]")
		}
		lines = append(lines, truncate(prefix+" "+hit.preview, w))
	}
	if len(m.searchResults) > max {
		lines = append(lines, dimTextStyle.Render(fmt.Sprintf("▼ %d more", len(m.searchResults)-max)))
	}
	return clipLines(header+"\n\n"+strings.Join(lines, "\n"), h)
}

// --- Help ---

func (m model) renderHelpPanel() string {
	h := m.bodyHeight()
	w := m.width
	lines := []string{
		titleStyle.Render("stickies — help"),
		"",
		dateHeaderStyle.Render("pages"),
		"  1 / 2 / 3       stickies / journal / search",
		"",
		dateHeaderStyle.Render("global"),
		"  n               new sticky",
		"  t               jump to today's journal entry",
		"  /               search",
		"  ?               toggle help",
		"  q · ctrl+c      quit",
		"",
		dateHeaderStyle.Render("list nav"),
		"  j/k · up/down   move cursor",
		"  g · G           top · bottom",
		"  ctrl+d · ctrl+u page down · up",
		"",
		dateHeaderStyle.Render("stickies"),
		"  E · enter       edit body inline",
		"  e               edit tags",
		"  p               toggle pin",
		"  o               open in $EDITOR",
		"  y               copy body",
		"  D               delete (confirm)",
		"",
		dateHeaderStyle.Render("journal"),
		"  E · enter       edit selected entry",
		"  t               jump to / create today",
		"  o               open in $EDITOR",
		"  y               copy body",
		"  D               delete (confirm)",
		"",
		dateHeaderStyle.Render("edit mode"),
		"  ctrl+s          save · close",
		"  ctrl+d          delete row",
		"  home / end      line start / end",
		"  ctrl+a / ctrl+e line start / end",
		"  alt+b / alt+f   word back / forward",
		"  alt+d / ctrl+w  delete word fwd / back",
		"  esc             cancel",
		"",
		dateHeaderStyle.Render("settings"),
		"  config dir: " + dimTextStyle.Render(m.dataDir),
		"  $EDITOR    " + dimTextStyle.Render("override external editor"),
		"  $STICKIES_DIR " + dimTextStyle.Render("override data directory"),
	}
	innerH := h - 2
	startIdx := m.helpScroll
	if startIdx >= len(lines) {
		startIdx = max0(0, len(lines)-innerH)
	}
	visible := lines[startIdx:]
	hasBottom := false
	if len(visible) > innerH {
		visible = visible[:innerH]
		hasBottom = true
	}
	hasTop := startIdx > 0
	body := strings.Join(visible, "\n")
	if hasTop {
		body = dimTextStyle.Render("▲ scroll up") + "\n" + body
	}
	if hasBottom {
		body = body + "\n" + dimTextStyle.Render("▼ scroll down")
	}
	return panelBox(false, w, h, "help", body)
}

// --- Layout helpers ---

// panelBox renders a bordered panel with an inner header label, sized to (w x h)
// where w/h include the borders.
func panelBox(active bool, w, h int, label, content string) string {
	style := panelStyle
	if active {
		style = panelActiveStyle
	}
	innerW := w - 4 // 2 borders + 2 padding
	if innerW < 4 {
		innerW = 4
	}
	innerH := h - 2 // 2 borders
	if innerH < 1 {
		innerH = 1
	}

	headerLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("105")).Render(label)
	// Reserve 1 line for label inside content area.
	bodyH := innerH - 1
	if bodyH < 1 {
		bodyH = 1
	}
	body := clipLines(content, bodyH)
	body = padToHeight(body, bodyH)
	combined := headerLabel + "\n" + body

	return style.Width(w - 2).Height(innerH).Render(combined)
}

func clipLines(s string, h int) string {
	if h <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}

func padToHeight(s string, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// wrapText word-wraps a body string to width w, preserving existing newlines.
// Falls back to character-break only for words longer than w.
func wrapText(s string, w int) string {
	if w <= 0 {
		return s
	}
	var out []string
	for _, line := range strings.Split(s, "\n") {
		out = append(out, wrapLineWords(line, w)...)
	}
	return strings.Join(out, "\n")
}

func wrapLineWords(line string, w int) []string {
	if line == "" {
		return []string{""}
	}
	if lipgloss.Width(line) <= w {
		return []string{line}
	}
	// Preserve leading whitespace of the original line.
	indent := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}
	words := strings.Fields(line)
	var out []string
	cur := indent
	curW := lipgloss.Width(indent)
	for _, word := range words {
		ww := lipgloss.Width(word)
		if ww > w {
			if cur != "" {
				out = append(out, cur)
				cur = ""
				curW = 0
			}
			rs := []rune(word)
			start := 0
			for start < len(rs) {
				end := start
				cw := 0
				for end < len(rs) {
					rw := lipgloss.Width(string(rs[end]))
					if cw+rw > w {
						break
					}
					cw += rw
					end++
				}
				if end == len(rs) {
					cur = string(rs[start:])
					curW = cw
					break
				}
				out = append(out, string(rs[start:end]))
				start = end
			}
			continue
		}
		sep := 0
		if cur != "" && cur != indent {
			sep = 1
		}
		if curW+sep+ww > w {
			out = append(out, cur)
			cur = word
			curW = ww
			continue
		}
		if sep == 1 {
			cur += " "
			curW++
		}
		cur += word
		curW += ww
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if len(r) > w-1 {
		return string(r[:w-1]) + "…"
	}
	return s
}

func max0(a, b int) int {
	if a > b {
		return a
	}
	return b
}
