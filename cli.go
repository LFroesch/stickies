package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
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
	}
	fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", cmd)
	return 2
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
