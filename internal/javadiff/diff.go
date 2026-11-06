// Package javadiff parses a unified diff (as GitHub serves a commit via
// "<commit url>.diff") and maps the changed line numbers of a .java file,
// on both sides of the change, to the methods that contain them, so a CVE
// fix commit can be turned into "which methods did this fix touch"
// without needing a real Java parser or a JDK.
package javadiff

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Change is how a file changed in the commit (CVE:F5).
type Change string

const (
	ChangeAdded    Change = "added"
	ChangeModified Change = "modified"
	ChangeDeleted  Change = "deleted"
	ChangeRenamed  Change = "renamed" // possibly with content changes too
)

// FileDiff is one file's changes within a unified diff, on both sides
// (CVE:F5): OldPath is the path before the commit ("" for an added file),
// NewPath the path after it ("" for a deleted file); OldLines are the
// pre-commit line numbers the diff removed, NewLines the post-commit line
// numbers it added. A modified line appears on both sides.
type FileDiff struct {
	OldPath  string
	NewPath  string
	Change   Change
	OldLines []int
	NewLines []int
}

// Path is the file's path after the commit, or before it for a deleted
// file.
func (f FileDiff) Path() string {
	if f.NewPath != "" {
		return f.NewPath
	}
	return f.OldPath
}

const hunkHeaderPrefix = "@@ -"

// ParseUnifiedDiff parses a multi-file git unified diff and returns one
// FileDiff per file, in diff order — including pure deletions, deleted
// files and renames without content changes. Hunk bodies are consumed by
// the line counts of their header, so a removed line whose content starts
// with "--" (shown as "--- ...") is never taken for a file header.
func ParseUnifiedDiff(diff string) ([]FileDiff, error) {
	var files []FileDiff
	var cur *FileDiff
	var oldLine, newLine, oldLeft, newLeft int
	inHunk := false

	flush := func() {
		if cur != nil {
			if cur.Change == "" {
				cur.Change = ChangeModified
				if cur.OldPath != "" && cur.NewPath != "" && cur.OldPath != cur.NewPath {
					cur.Change = ChangeRenamed
				}
			}
			files = append(files, *cur)
		}
		cur = nil
		inHunk = false
	}

	sc := bufio.NewScanner(strings.NewReader(diff))
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()

		if inHunk && (oldLeft > 0 || newLeft > 0) {
			switch {
			case strings.HasPrefix(line, "+"):
				cur.NewLines = append(cur.NewLines, newLine)
				newLine++
				newLeft--
			case strings.HasPrefix(line, "-"):
				cur.OldLines = append(cur.OldLines, oldLine)
				oldLine++
				oldLeft--
			case strings.HasPrefix(line, "\\"):
				// "\ No newline at end of file" — not a content line.
			default:
				// A context line (starts with a space, or is blank).
				oldLine++
				newLine++
				oldLeft--
				newLeft--
			}
			continue
		}
		inHunk = false

		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			a, b := parseGitHeader(strings.TrimPrefix(line, "diff --git "))
			cur = &FileDiff{OldPath: a, NewPath: b}

		case cur == nil && (strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ")):
			// A diff without "diff --git" headers (plain unified diff).
			cur = &FileDiff{}
			fallthrough

		case cur != nil && strings.HasPrefix(line, "--- "):
			if p, ok := sidePath(strings.TrimPrefix(line, "--- "), "a/"); ok {
				cur.OldPath = p
			} else {
				cur.OldPath, cur.Change = "", ChangeAdded
			}
			if strings.HasPrefix(line, "+++ ") { // reached by fallthrough
				setNewPath(cur, line)
			}

		case cur != nil && strings.HasPrefix(line, "+++ "):
			setNewPath(cur, line)

		case cur != nil && strings.HasPrefix(line, "new file mode"):
			cur.Change, cur.OldPath = ChangeAdded, ""
		case cur != nil && strings.HasPrefix(line, "deleted file mode"):
			cur.Change, cur.NewPath = ChangeDeleted, ""
		case cur != nil && strings.HasPrefix(line, "rename from "):
			cur.OldPath, cur.Change = strings.TrimPrefix(line, "rename from "), ChangeRenamed
		case cur != nil && strings.HasPrefix(line, "rename to "):
			cur.NewPath, cur.Change = strings.TrimPrefix(line, "rename to "), ChangeRenamed

		case cur != nil && strings.HasPrefix(line, hunkHeaderPrefix):
			h, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("parsing hunk header %q: %w", line, err)
			}
			oldLine, oldLeft, newLine, newLeft = h[0], h[1], h[2], h[3]
			inHunk = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading diff: %w", err)
	}
	flush()
	return files, nil
}

func setNewPath(cur *FileDiff, line string) {
	if p, ok := sidePath(strings.TrimPrefix(line, "+++ "), "b/"); ok {
		cur.NewPath = p
	} else {
		cur.NewPath, cur.Change = "", ChangeDeleted
	}
}

// sidePath strips the "a/"/"b/" prefix and a trailing tab-separated
// timestamp; ok is false for /dev/null.
func sidePath(p, prefix string) (string, bool) {
	if i := strings.IndexByte(p, '\t'); i >= 0 {
		p = p[:i]
	}
	if p == "/dev/null" {
		return "", false
	}
	return strings.TrimPrefix(p, prefix), true
}

// parseGitHeader splits "a/<old> b/<new>". Paths with spaces are
// ambiguous in this line; the "---"/"+++"/"rename" lines that follow
// override it whenever they are present.
func parseGitHeader(s string) (string, string) {
	if i := strings.Index(s, " b/"); i >= 0 && strings.HasPrefix(s, "a/") {
		return s[2:i], s[i+3:]
	}
	return "", ""
}

// parseHunkHeader extracts oldStart, oldLines, newStart, newLines from
// "@@ -oldStart[,oldLines] +newStart[,newLines] @@ ..." (an omitted count
// is 1).
func parseHunkHeader(header string) ([4]int, error) {
	var out [4]int
	rest := strings.TrimPrefix(header, hunkHeaderPrefix)
	end := strings.Index(rest, " @@")
	if end < 0 {
		return out, fmt.Errorf("no closing @@")
	}
	oldR, newR, ok := strings.Cut(rest[:end], " +")
	if !ok {
		return out, fmt.Errorf("no '+' range")
	}
	for i, r := range []string{oldR, newR} {
		start, count, hasCount := strings.Cut(r, ",")
		s, err := strconv.Atoi(start)
		if err != nil {
			return out, err
		}
		c := 1
		if hasCount {
			if c, err = strconv.Atoi(count); err != nil {
				return out, err
			}
		}
		out[2*i], out[2*i+1] = s, c
	}
	return out, nil
}
