package main

import "testing"

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
