package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDiscardPendingNewStickyRemovesUnsavedPlaceholder(t *testing.T) {
	m := model{
		stickies: []Sticky{
			{ID: "keep-1"},
			{ID: "new-note"},
			{ID: "keep-2"},
		},
		stickyCursor:   1,
		editingSticky:  1,
		creatingSticky: true,
	}

	m.discardPendingNewSticky()

	if len(m.stickies) != 2 {
		t.Fatalf("expected 2 stickies after discard, got %d", len(m.stickies))
	}
	if m.stickies[0].ID != "keep-1" || m.stickies[1].ID != "keep-2" {
		t.Fatalf("unexpected stickies after discard: %+v", m.stickies)
	}
	if m.stickyCursor != 1 {
		t.Fatalf("expected cursor to clamp to remaining item, got %d", m.stickyCursor)
	}
}

func TestHandleJournalKeyCtrlDAndCtrlUPageByVisibleBody(t *testing.T) {
	m := model{
		height:        24,
		journal:       make([]JournalEntry, 20),
		journalCursor: 0,
	}

	next, _ := m.handleJournalKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	gotDown := next.(model).journalCursor
	if gotDown != 10 {
		t.Fatalf("expected ctrl+d to move journal cursor by 10, got %d", gotDown)
	}

	m.journalCursor = 15
	prev, _ := m.handleJournalKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	gotUp := prev.(model).journalCursor
	if gotUp != 5 {
		t.Fatalf("expected ctrl+u to move journal cursor back by 10, got %d", gotUp)
	}
}

func TestHandleStickiesKeyCtrlDUsesDynamicPageStep(t *testing.T) {
	m := model{
		height:       18,
		stickies:     make([]Sticky, 20),
		stickyCursor: 2,
	}

	next, _ := m.handleStickiesKey(tea.KeyMsg{Type: tea.KeyCtrlD})
	got := next.(model).stickyCursor
	if got != 9 {
		t.Fatalf("expected ctrl+d to move sticky cursor by 7, got %d", got)
	}
}

func TestDeleteCurrentBodyRowRemovesActiveMiddleLine(t *testing.T) {
	body := textarea.New()
	body.SetValue("first\nsecond\nthird")
	body.CursorUp()
	body.CursorStart()

	m := model{bodyArea: body}
	m.deleteCurrentBodyRow()

	if got := m.bodyArea.Value(); got != "first\nthird" {
		t.Fatalf("expected middle row to be deleted, got %q", got)
	}
	if got := m.bodyArea.Line(); got != 1 {
		t.Fatalf("expected cursor to land on replacement row, got line %d", got)
	}
}

func TestDeleteCurrentBodyRowClearsOnlyLine(t *testing.T) {
	body := textarea.New()
	body.SetValue("only")

	m := model{bodyArea: body}
	m.deleteCurrentBodyRow()

	if got := m.bodyArea.Value(); got != "" {
		t.Fatalf("expected only row to be deleted, got %q", got)
	}
	if got := m.bodyArea.Line(); got != 0 {
		t.Fatalf("expected cursor to reset to first line, got line %d", got)
	}
}
