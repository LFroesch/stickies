package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	stickiesSubdir = "stickies"
	journalSubdir  = "journal"
	dateLayout     = "2006-01-02"
	timeLayout     = time.RFC3339
)

func ensureDirs(dir string) error {
	for _, sub := range []string{stickiesSubdir, journalSubdir} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			return err
		}
	}
	return nil
}

func newID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// --- Frontmatter ---
//
// File format:
//   ---
//   key: value
//   key: value
//   ---
//   <body>
//
// Values are strings; lists encoded as comma-separated. Tiny inline parser to
// avoid pulling in a YAML dependency for what is effectively a flat key/value
// header.

func parseFrontmatter(raw string) (map[string]string, string) {
	meta := map[string]string{}
	if !strings.HasPrefix(raw, "---\n") && !strings.HasPrefix(raw, "---\r\n") {
		return meta, raw
	}
	rest := strings.TrimPrefix(raw, "---\n")
	rest = strings.TrimPrefix(rest, "---\r\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return meta, raw
	}
	header := rest[:end]
	body := rest[end+len("\n---"):]
	body = strings.TrimPrefix(body, "\r")
	body = strings.TrimPrefix(body, "\n")
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		meta[key] = val
	}
	return meta, body
}

func writeFrontmatter(meta map[string]string, body string) string {
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("---\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(meta[k])
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	b.WriteString(body)
	return b.String()
}

func atomicWrite(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// --- Stickies ---

func stickyPath(dir, id string) string {
	return filepath.Join(dir, stickiesSubdir, id+".md")
}

func loadStickies(dir string) ([]Sticky, error) {
	matches, err := filepath.Glob(filepath.Join(dir, stickiesSubdir, "*.md"))
	if err != nil {
		return nil, err
	}
	out := make([]Sticky, 0, len(matches))
	for _, p := range matches {
		s, err := readSticky(p)
		if err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func readSticky(path string) (Sticky, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Sticky{}, err
	}
	meta, body := parseFrontmatter(string(raw))
	id := meta["id"]
	if id == "" {
		id = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	s := Sticky{
		ID:     id,
		Body:   strings.TrimRight(body, "\n"),
		Pinned: meta["pinned"] == "true",
	}
	if t := meta["tags"]; t != "" {
		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				s.Tags = append(s.Tags, tag)
			}
		}
	}
	if c, err := time.Parse(timeLayout, meta["created"]); err == nil {
		s.Created = c
	}
	if u, err := time.Parse(timeLayout, meta["updated"]); err == nil {
		s.Updated = u
	}
	return s, nil
}

func saveSticky(dir string, s Sticky) error {
	meta := map[string]string{
		"id":      s.ID,
		"pinned":  fmt.Sprintf("%t", s.Pinned),
		"tags":    strings.Join(s.Tags, ", "),
		"created": s.Created.Format(timeLayout),
		"updated": s.Updated.Format(timeLayout),
	}
	body := s.Body
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return atomicWrite(stickyPath(dir, s.ID), writeFrontmatter(meta, body))
}

func deleteSticky(dir, id string) error {
	return os.Remove(stickyPath(dir, id))
}

func sortStickies(s []Sticky) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].Pinned != s[j].Pinned {
			return s[i].Pinned
		}
		return s[i].Updated.After(s[j].Updated)
	})
}

// --- Journal ---

func journalPath(dir, date string) string {
	return filepath.Join(dir, journalSubdir, date+".md")
}

func loadJournal(dir string) ([]JournalEntry, error) {
	matches, err := filepath.Glob(filepath.Join(dir, journalSubdir, "*.md"))
	if err != nil {
		return nil, err
	}
	out := make([]JournalEntry, 0, len(matches))
	for _, p := range matches {
		date := strings.TrimSuffix(filepath.Base(p), ".md")
		if _, err := time.Parse(dateLayout, date); err != nil {
			continue
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		meta, body := parseFrontmatter(string(raw))
		e := JournalEntry{
			Date: date,
			Body: strings.TrimRight(body, "\n"),
		}
		if u, err := time.Parse(timeLayout, meta["updated"]); err == nil {
			e.Updated = u
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date > out[j].Date })
	return out, nil
}

func saveJournal(dir string, e JournalEntry) error {
	meta := map[string]string{
		"date":    e.Date,
		"updated": e.Updated.Format(timeLayout),
	}
	body := e.Body
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return atomicWrite(journalPath(dir, e.Date), writeFrontmatter(meta, body))
}

func deleteJournal(dir, date string) error {
	return os.Remove(journalPath(dir, date))
}

func todayDate() string  { return time.Now().Format(dateLayout) }
func yesterdayDate() string {
	return time.Now().AddDate(0, 0, -1).Format(dateLayout)
}

// findJournal returns index of date in journal slice, or -1.
func findJournal(j []JournalEntry, date string) int {
	for i, e := range j {
		if e.Date == date {
			return i
		}
	}
	return -1
}

// upsertJournal inserts or updates an entry, preserving newest-first order.
func upsertJournal(j []JournalEntry, e JournalEntry) []JournalEntry {
	if i := findJournal(j, e.Date); i >= 0 {
		j[i] = e
		return j
	}
	j = append(j, e)
	sort.Slice(j, func(i, k int) bool { return j[i].Date > j[k].Date })
	return j
}

// --- Search ---

func searchAll(stickies []Sticky, journal []JournalEntry, q string) []searchHit {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	var hits []searchHit
	for i, s := range stickies {
		if matchSticky(s, q) {
			hits = append(hits, searchHit{
				stickyIdx: i,
				preview:   firstLine(s.Body),
			})
		}
	}
	for _, e := range journal {
		if strings.Contains(strings.ToLower(e.Body), q) || strings.Contains(e.Date, q) {
			hits = append(hits, searchHit{
				isJournal: true,
				date:      e.Date,
				preview:   firstLine(e.Body),
			})
		}
	}
	return hits
}

func matchSticky(s Sticky, q string) bool {
	if strings.Contains(strings.ToLower(s.Body), q) {
		return true
	}
	for _, t := range s.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return s
}

// --- External editor ---

func resolveEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "vi"
}

// --- Tag parsing ---

func parseTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// --- Misc ---

func clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
