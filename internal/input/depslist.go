package input

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ansiEscape strips ANSI color codes, which `mvn dependency:list` emits
// whenever Maven's color output is on (observed on real builds; the plan
// anticipated this).
var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

// ParseDepsList parses the output of
// `mvn dependency:list -DincludeScope=runtime -DoutputFile=deps.txt`.
//
// Each dependency line looks like
// "   group:artifact:type[:classifier]:version:scope[ -- module x[ (auto)|[auto]]]"
// Only the first whitespace-delimited field is the Maven coordinate; any
// ANSI color codes and any text after it (module/classifier/optional
// annotations appended by newer plugin versions) are ignored. Lines that
// don't parse as a 4- or 5-part coordinate (headers, blank lines, "None")
// are skipped rather than treated as errors, since the exact wording of
// non-coordinate lines varies across Maven versions.
func ParseDepsList(r io.Reader) ([]GAV, error) {
	var gavs []GAV
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := ansiEscape.ReplaceAllString(sc.Text(), "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		field := strings.Fields(line)[0]
		gav, ok := parseCoordinate(field)
		if !ok {
			continue
		}
		gavs = append(gavs, gav)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("parsing deps list: %w", err)
	}
	return gavs, nil
}

// parseCoordinate parses a single "group:artifact:type[:classifier]:version:scope"
// token. ok is false if it doesn't have the right shape (e.g. a header line's
// first word).
func parseCoordinate(s string) (GAV, bool) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 5:
		// group:artifact:type:version:scope
		return GAV{Group: parts[0], Artifact: parts[1], Version: parts[3]}, true
	case 6:
		// group:artifact:type:classifier:version:scope
		return GAV{Group: parts[0], Artifact: parts[1], Classifier: parts[3], Version: parts[4]}, true
	default:
		return GAV{}, false
	}
}
