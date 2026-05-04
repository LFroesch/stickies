package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnsureDirsCreatesSubdirs(t *testing.T) {
	dir := t.TempDir()
	if err := ensureDirs(dir); err != nil {
		t.Fatalf("ensureDirs: %v", err)
	}
	for _, sub := range []string{stickiesSubdir, journalSubdir} {
		if _, err := os.Stat(filepath.Join(dir, sub)); err != nil {
			t.Errorf("missing subdir %s: %v", sub, err)
		}
	}
}

func TestSortStickiesPinnedFirstThenRecent(t *testing.T) {
	now := time.Now()
	s := []Sticky{
		{ID: "old-unpinned", Updated: now.Add(-2 * time.Hour)},
		{ID: "new-unpinned", Updated: now},
		{ID: "old-pinned", Pinned: true, Updated: now.Add(-3 * time.Hour)},
		{ID: "new-pinned", Pinned: true, Updated: now.Add(-time.Hour)},
	}
	sortStickies(s)
	want := []string{"new-pinned", "old-pinned", "new-unpinned", "old-unpinned"}
	for i, id := range want {
		if s[i].ID != id {
			t.Errorf("position %d: want %s, got %s", i, id, s[i].ID)
		}
	}
}

func TestParseFrontmatterRoundTrip(t *testing.T) {
	body := "first line\nsecond line\n"
	meta := map[string]string{"id": "abc", "tags": "a,b,c"}
	raw := writeFrontmatter(meta, body)
	if !strings.HasPrefix(raw, "---\n") {
		t.Fatalf("expected frontmatter prefix, got %q", raw[:10])
	}
	gotMeta, gotBody := parseFrontmatter(raw)
	if gotMeta["id"] != "abc" || gotMeta["tags"] != "a,b,c" {
		t.Errorf("meta mismatch: %+v", gotMeta)
	}
	if gotBody != body {
		t.Errorf("body mismatch: want %q, got %q", body, gotBody)
	}
}

func TestParseFrontmatterNoHeader(t *testing.T) {
	raw := "no header here\nplain body\n"
	meta, body := parseFrontmatter(raw)
	if len(meta) != 0 {
		t.Errorf("expected empty meta, got %+v", meta)
	}
	if body != raw {
		t.Errorf("body should equal raw input")
	}
}

func TestSplitCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "plain editor",
			input: "nvim",
			want:  []string{"nvim"},
		},
		{
			name:  "quoted args",
			input: `code --wait --reuse-window "/tmp/my notes.md"`,
			want:  []string{"code", "--wait", "--reuse-window", "/tmp/my notes.md"},
		},
		{
			name:  "escaped spaces",
			input: `emacsclient --socket-name default /tmp/my\ notes.md`,
			want:  []string{"emacsclient", "--socket-name", "default", "/tmp/my notes.md"},
		},
		{
			name:    "unterminated quote",
			input:   `code --wait "abc`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommandLine(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got args %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitCommandLine: %v", err)
			}
			if strings.Join(got, "\x00") != strings.Join(tt.want, "\x00") {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}
