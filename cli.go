package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// CLI subcommands for piping into other tools.
//
// Usage:
//   stickies ls                  list stickies (id<TAB>title)
//   stickies cat <q>             print sticky body (match by id-prefix or title)
//   stickies today               print today's journal body
//   stickies day <YYYY-MM-DD>    print specified day's journal body
//   stickies days                list journal dates
//   stickies search <query>      print matches across stickies + journal
//   stickies new [text]          create sticky from arg or stdin
//   stickies log [text]          append timestamped line to today's journal

func runCLI(cmd string, args []string, dataDir string, out io.Writer) int {
	switch cmd {
	case "ls", "list":
		return cmdLs(dataDir, out)
	case "cat", "show", "get":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: stickies cat <id-or-title>")
			return 2
		}
		return cmdCat(dataDir, args[0], out)
	case "today":
		return cmdJournal(dataDir, todayDate(), out)
	case "day":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: stickies day <YYYY-MM-DD>")
			return 2
		}
		return cmdJournal(dataDir, args[0], out)
	case "days":
		return cmdDays(dataDir, out)
	case "search", "grep":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: stickies search <query>")
			return 2
		}
		return cmdSearch(dataDir, strings.Join(args, " "), out)
	case "new", "n":
		body, err := argOrStdin(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "usage: stickies new <text>   (or pipe via stdin)")
			return 2
		}
		return cmdNew(dataDir, body, out)
	case "log", "l":
		body, err := argOrStdin(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "usage: stickies log <text>   (or pipe via stdin)")
			return 2
		}
		return cmdLog(dataDir, body, out)
	}
	fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", cmd)
	return 2
}

// argOrStdin returns args joined with spaces, or stdin if args is empty and
// stdin is piped. Returns an error when neither is available.
func argOrStdin(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		return "", errors.New("no input")
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func cmdLs(dataDir string, out io.Writer) int {
	s, err := loadStickies(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	sortStickies(s)
	for _, st := range s {
		title := firstLine(st.Body)
		if title == "" {
			title = "(empty)"
		}
		mark := " "
		if st.Pinned {
			mark = "*"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\n", shortID(st.ID), mark, title)
	}
	return 0
}

func cmdCat(dataDir, query string, out io.Writer) int {
	s, err := loadStickies(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	hit := lookupSticky(s, query)
	if hit < 0 {
		fmt.Fprintf(os.Stderr, "no sticky matching %q\n", query)
		return 1
	}
	fmt.Fprintln(out, s[hit].Body)
	return 0
}

func cmdJournal(dataDir, date string, out io.Writer) int {
	j, err := loadJournal(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	i := findJournal(j, date)
	if i < 0 {
		fmt.Fprintf(os.Stderr, "no journal entry for %s\n", date)
		return 1
	}
	fmt.Fprintln(out, j[i].Body)
	return 0
}

func cmdDays(dataDir string, out io.Writer) int {
	j, err := loadJournal(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, e := range j {
		fmt.Fprintln(out, e.Date)
	}
	return 0
}

func cmdSearch(dataDir, query string, out io.Writer) int {
	s, err := loadStickies(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	j, err := loadJournal(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	hits := searchAll(s, j, query)
	for _, h := range hits {
		if h.isJournal {
			fmt.Fprintf(out, "journal\t%s\t%s\n", h.date, h.preview)
		} else if h.stickyIdx >= 0 && h.stickyIdx < len(s) {
			fmt.Fprintf(out, "sticky\t%s\t%s\n", shortID(s[h.stickyIdx].ID), h.preview)
		}
	}
	return 0
}

func cmdNew(dataDir, body string, out io.Writer) int {
	body = strings.TrimRight(body, "\n")
	if strings.TrimSpace(body) == "" {
		fmt.Fprintln(os.Stderr, "stickies new: empty body")
		return 2
	}
	now := time.Now()
	s := Sticky{
		ID:      newID(),
		Body:    body,
		Created: now,
		Updated: now,
	}
	if err := saveSticky(dataDir, s); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(out, "%s\t%s\n", shortID(s.ID), firstLine(s.Body))
	return 0
}

func cmdLog(dataDir, line string, out io.Writer) int {
	line = strings.TrimRight(line, "\n")
	if strings.TrimSpace(line) == "" {
		fmt.Fprintln(os.Stderr, "stickies log: empty entry")
		return 2
	}
	j, err := loadJournal(dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	date := todayDate()
	now := time.Now()
	entry := fmt.Sprintf("- %s %s", now.Format("15:04"), line)
	body := entry
	if i := findJournal(j, date); i >= 0 && j[i].Body != "" {
		body = j[i].Body + "\n" + entry
	}
	e := JournalEntry{Date: date, Body: body, Updated: now}
	if err := saveJournal(dataDir, e); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintln(out, entry)
	return 0
}

// shortID returns the first 8 chars of a hex ID — enough to be unique in
// practice and short enough to type/pipe.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// lookupSticky resolves a query (id, id-prefix, or case-insensitive title
// substring) to an index in s, or -1.
func lookupSticky(s []Sticky, q string) int {
	q = strings.TrimSpace(q)
	if q == "" {
		return -1
	}
	// exact id
	for i, st := range s {
		if st.ID == q {
			return i
		}
	}
	// id prefix (unambiguous)
	var prefixHits []int
	for i, st := range s {
		if strings.HasPrefix(st.ID, q) {
			prefixHits = append(prefixHits, i)
		}
	}
	if len(prefixHits) == 1 {
		return prefixHits[0]
	}
	// title substring (case-insensitive); prefer earliest pinned/recent
	ql := strings.ToLower(q)
	var titleHits []int
	for i, st := range s {
		if strings.Contains(strings.ToLower(firstLine(st.Body)), ql) {
			titleHits = append(titleHits, i)
		}
	}
	if len(titleHits) == 0 {
		return -1
	}
	sort.SliceStable(titleHits, func(a, b int) bool {
		ia, ib := titleHits[a], titleHits[b]
		if s[ia].Pinned != s[ib].Pinned {
			return s[ia].Pinned
		}
		return s[ia].Updated.After(s[ib].Updated)
	})
	return titleHits[0]
}
