// Package javadiff parses a unified diff (as GitHub serves a commit via
// "<commit url>.diff") and maps the new-side changed line numbers of a
// .java file to the methods that contain them, so a CVE fix commit can be
// turned into "which methods did this fix touch" without needing a real
// Java parser or a JDK.
package javadiff

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// FileDiff is one file's changes within a unified diff: its path (as it
// exists after the commit, i.e. the "+++ b/..." side) and every line
// number, in the post-commit file, that the diff added or changed.
type FileDiff struct {
	Path         string
	ChangedLines []int
}

// hunkHeader matches "@@ -oldStart[,oldLines] +newStart[,newLines] @@...".
const hunkHeaderPrefix = "@@ -"

// ParseUnifiedDiff parses a multi-file unified diff and returns one
// FileDiff per file that has at least one added/changed line. Pure
// deletions (a hunk with only '-' lines) contribute no changed lines,
// since there's nothing on the new side to map to a method.
func ParseUnifiedDiff(diff string) ([]FileDiff, error) {
	var files []FileDiff
	var cur *FileDiff
	newLine := 0 // the next new-file line number a ' '/'+' line will have

	flush := func() {
		if cur != nil && len(cur.ChangedLines) > 0 {
			files = append(files, *cur)
		}
		cur = nil
	}

	sc := bufio.NewScanner(strings.NewReader(diff))
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Text()

		switch {
		case strings.HasPrefix(line, "+++ "):
			flush()
			path := strings.TrimPrefix(line, "+++ ")
			path = strings.TrimPrefix(path, "b/")
			// A tab sometimes separates the path from a trailing
			// timestamp; GitHub doesn't emit that, but be tolerant.
			if i := strings.IndexByte(path, '\t'); i >= 0 {
				path = path[:i]
			}
			if path == "/dev/null" {
				cur = nil // file was deleted; nothing to map on the new side
				continue
			}
			cur = &FileDiff{Path: path}

		case strings.HasPrefix(line, hunkHeaderPrefix):
			start, err := parseHunkNewStart(line)
			if err != nil {
				return nil, fmt.Errorf("parsing hunk header %q: %w", line, err)
			}
			newLine = start

		case cur == nil:
			// Outside any file we're tracking (e.g. the "diff --git"/
			// "index .." preamble lines, or a file we decided to skip).
			continue

		case strings.HasPrefix(line, "+"):
			cur.ChangedLines = append(cur.ChangedLines, newLine)
			newLine++

		case strings.HasPrefix(line, "-"):
			// A removed line doesn't exist in the new file, so it
			// doesn't advance newLine and isn't itself a changed line
			// to map (there's nothing there to map it to).

		case strings.HasPrefix(line, "\\"):
			// "\ No newline at end of file" — not a content line.

		default:
			// A context line (starts with a space, or is blank).
			newLine++
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading diff: %w", err)
	}
	flush()
	return files, nil
}

// parseHunkNewStart extracts newStart from a hunk header
// "@@ -oldStart,oldLines +newStart,newLines @@ ...".
func parseHunkNewStart(header string) (int, error) {
	plus := strings.IndexByte(header, '+')
	if plus < 0 {
		return 0, fmt.Errorf("no '+' field")
	}
	rest := header[plus+1:]
	end := strings.IndexAny(rest, ", @")
	if end < 0 {
		return 0, fmt.Errorf("malformed new-range field")
	}
	return strconv.Atoi(rest[:end])
}
