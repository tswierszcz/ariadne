// Package mavenversion compares Maven versions and evaluates version
// ranges (CVE:F4). The ordering follows Maven's documented "Version Order
// Specification" (https://maven.apache.org/pom.html#version-order-specification),
// implemented here from that specification — not a SemVer comparator:
//
//   - a version is split into tokens at '.', '-', and at every transition
//     between digits and letters ("1.0alpha1" → 1, 0, alpha, 1); a '-'
//     (or a digit/letter transition) starts a sub-list, a '.' does not;
//   - trailing "null" tokens (0, "", "final", "ga", "release") are
//     ignored, so 1 = 1.0 = 1.0.0 = 1-ga = 1.final;
//   - numbers compare numerically; qualifiers order alpha < beta <
//     milestone < rc = cr < snapshot < "" = final = ga = release < sp,
//     with a, b, m as one-letter aliases of alpha, beta, milestone when
//     directly followed by a digit; any other qualifier sorts after sp,
//     lexically (case-insensitive);
//   - a number sorts after a qualifier at the same position, and a
//     sub-list after a qualifier but before a number.
//
// Comparing two versions never fails. Deciding whether a version is in a
// range is a different question: Regular reports versions for which that
// decision is trusted; anything else (SNAPSHOT builds, versions that do
// not start with a number, unusual characters) is undetermined, never
// decided by a textual fallback.
package mavenversion

import (
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

type kind int

const (
	kNumber kind = iota
	kString
	kList
)

type item struct {
	kind kind
	num  *big.Int
	str  string // normalized qualifier
	list []item
}

var qualifierOrder = map[string]int{"alpha": 0, "beta": 1, "milestone": 2, "rc": 3, "snapshot": 4, "": 5, "sp": 6}

func normalizeQualifier(s string, followedByDigit bool) string {
	s = strings.ToLower(s)
	if followedByDigit {
		switch s {
		case "a":
			return "alpha"
		case "b":
			return "beta"
		case "m":
			return "milestone"
		}
	}
	switch s {
	case "cr":
		return "rc"
	case "ga", "final", "release":
		return ""
	}
	return s
}

func isNull(it item) bool {
	switch it.kind {
	case kNumber:
		return it.num.Sign() == 0
	case kString:
		return it.str == ""
	}
	return len(it.list) == 0
}

// parse tokenizes v into Maven's nested item list (Maven 3.9 rules,
// checked token-for-token against Maven 3.9.12): '.' separates items of
// the current list; '-' (on a non-empty list), every digit/letter
// transition and a qualifier following other items start a new sub-list;
// each list then drops trailing null items.
func parse(v string) []item {
	v = strings.ToLower(v)
	type frame struct{ list []item }
	stack := []*frame{{}}
	cur := func() *frame { return stack[len(stack)-1] }
	newList := func() { stack = append(stack, &frame{}) }
	add := func(it item) {
		if it.kind == kString {
			if len(cur().list) > 0 {
				newList() // ".X" is treated as "-X" for a string qualifier X
			}
		}
		cur().list = append(cur().list, it)
	}
	token := func(digit bool, s string, followedByDigit bool) item {
		if digit {
			n := new(big.Int)
			n.SetString(s, 10)
			return item{kind: kNumber, num: n}
		}
		return item{kind: kString, str: normalizeQualifier(s, followedByDigit)}
	}
	isDigit, start := false, 0
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '.' || c == '-':
			if i == start {
				add(item{kind: kNumber, num: new(big.Int)})
			} else {
				add(token(isDigit, v[start:i], false))
			}
			start = i + 1
			if c == '-' && len(cur().list) > 0 {
				newList()
			}
		case c >= '0' && c <= '9':
			if !isDigit && i > start {
				add(token(false, v[start:i], true))
				start = i
				newList()
			}
			isDigit = true
		default:
			if isDigit && i > start {
				add(token(true, v[start:i], false))
				start = i
				newList()
			}
			isDigit = false
		}
	}
	if len(v) > start {
		add(token(isDigit, v[start:], false))
	}
	// Close every open list innermost first, normalizing each.
	for len(stack) > 1 {
		inner := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		add(item{kind: kList, list: normalize(inner.list)})
	}
	return normalize(stack[0].list)
}

// normalize removes trailing null items, continuing past non-null
// sub-lists and stopping at the first other non-null item (Maven's
// ListItem.normalize).
func normalize(l []item) []item {
	for i := len(l) - 1; i >= 0; i-- {
		if isNull(l[i]) {
			l = append(l[:i], l[i+1:]...)
		} else if l[i].kind != kList {
			break
		}
	}
	return l
}

// cmp compares item a with b; b may be nil (absent: compared with null).
func cmp(a item, b *item) int {
	switch a.kind {
	case kNumber:
		if b == nil {
			if a.num.Sign() == 0 {
				return 0
			}
			return 1
		}
		switch b.kind {
		case kNumber:
			return a.num.Cmp(b.num)
		default:
			return 1 // number > qualifier, number > list
		}
	case kString:
		if b == nil {
			return cmpQualifier(a.str, "")
		}
		switch b.kind {
		case kNumber, kList:
			return -1 // qualifier < number, qualifier < list
		default:
			return cmpQualifier(a.str, b.str)
		}
	}
	// list
	if b == nil {
		for _, it := range a.list {
			if r := cmp(it, nil); r != 0 {
				return r
			}
		}
		return 0
	}
	switch b.kind {
	case kNumber:
		return -1 // list < number
	case kString:
		return 1 // list > qualifier
	}
	return cmpLists(a.list, b.list)
}

func cmpLists(a, b []item) int {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		var r int
		switch {
		case i >= len(a):
			r = -cmp(b[i], nil)
		case i >= len(b):
			r = cmp(a[i], nil)
		default:
			r = cmp(a[i], &b[i])
		}
		if r != 0 {
			return r
		}
	}
	return 0
}

func cmpQualifier(a, b string) int {
	ra, ka := qualifierOrder[a]
	rb, kb := qualifierOrder[b]
	switch {
	case ka && kb:
		return sign(ra - rb)
	case ka:
		return -1 // known qualifiers sort before unknown ones
	case kb:
		return 1
	}
	return strings.Compare(a, b)
}

func sign(x int) int {
	switch {
	case x < 0:
		return -1
	case x > 0:
		return 1
	}
	return 0
}

// Compare returns -1, 0 or 1 as a is older than, equal to or newer than b.
func Compare(a, b string) int {
	return cmpLists(parse(a), parse(b))
}

var regular = regexp.MustCompile(`^[0-9]+([.\-_+]?[0-9A-Za-z]+)*$`)

// Regular reports whether range decisions for v are trusted: it starts
// with a number, uses only letters, digits and . - _ +, and is not a
// SNAPSHOT build.
func Regular(v string) bool {
	return regular.MatchString(v) && !strings.Contains(strings.ToUpper(v), "SNAPSHOT")
}

// Decision is the outcome of a range check.
type Decision int

const (
	Undetermined Decision = iota
	In
	Out
)

func (d Decision) String() string {
	return [...]string{"undetermined", "in", "out"}[d]
}

// bound is one end of an interval; empty version = unbounded.
type bound struct {
	version   string
	inclusive bool
}

type interval struct{ lo, hi bound }

// Range is a Maven version range: a union of intervals.
type Range struct {
	spec      string
	intervals []interval
}

func (r Range) String() string { return r.spec }

// ParseRange parses Maven range syntax: "[1.0,2.0)", "(,1.0]", "[1.5]",
// "[1.0,)", and comma-separated unions "[1,2),[3,4)". A bare version
// ("1.0") is Maven's soft requirement and is rejected here: a mapping
// must say which versions it covers.
func ParseRange(spec string) (Range, error) {
	s := strings.ReplaceAll(spec, " ", "")
	if s == "" {
		return Range{}, fmt.Errorf("empty version range")
	}
	r := Range{spec: spec}
	for s != "" {
		if s[0] != '[' && s[0] != '(' {
			return Range{}, fmt.Errorf("version range %q: expected '[' or '(' (a bare version is not a range)", spec)
		}
		end := strings.IndexAny(s, "])")
		if end < 0 {
			return Range{}, fmt.Errorf("version range %q: unclosed interval", spec)
		}
		body := s[1:end]
		iv := interval{lo: bound{inclusive: s[0] == '['}, hi: bound{inclusive: s[end] == ']'}}
		if lo, hi, comma := strings.Cut(body, ","); !comma {
			if !iv.lo.inclusive || !iv.hi.inclusive || body == "" {
				return Range{}, fmt.Errorf("version range %q: a single version must be written [v]", spec)
			}
			iv.lo.version, iv.hi.version = body, body
		} else {
			if strings.Contains(hi, ",") {
				return Range{}, fmt.Errorf("version range %q: too many commas in one interval", spec)
			}
			iv.lo.version, iv.hi.version = lo, hi
			if (lo == "" && iv.lo.inclusive) || (hi == "" && iv.hi.inclusive) {
				return Range{}, fmt.Errorf("version range %q: an unbounded side must be open", spec)
			}
			if lo != "" && hi != "" && Compare(lo, hi) > 0 {
				return Range{}, fmt.Errorf("version range %q: lower bound above upper bound", spec)
			}
		}
		for _, v := range []string{iv.lo.version, iv.hi.version} {
			if v != "" && !Regular(v) {
				return Range{}, fmt.Errorf("version range %q: bound %q is not a regular version", spec, v)
			}
		}
		r.intervals = append(r.intervals, iv)
		s = s[end+1:]
		if strings.HasPrefix(s, ",") {
			s = s[1:]
			if s == "" {
				return Range{}, fmt.Errorf("version range %q: trailing comma", spec)
			}
		} else if s != "" {
			return Range{}, fmt.Errorf("version range %q: expected ',' between intervals", spec)
		}
	}
	return r, nil
}

// Contains decides whether v is in r; Undetermined when v is not Regular.
func (r Range) Contains(v string) Decision {
	if !Regular(v) {
		return Undetermined
	}
	for _, iv := range r.intervals {
		if iv.lo.version != "" {
			c := Compare(v, iv.lo.version)
			if c < 0 || (c == 0 && !iv.lo.inclusive) {
				continue
			}
		}
		if iv.hi.version != "" {
			c := Compare(v, iv.hi.version)
			if c > 0 || (c == 0 && !iv.hi.inclusive) {
				continue
			}
		}
		return In
	}
	return Out
}

// EvalOSV decides v against OSV "ECOSYSTEM" range events for Maven
// (introduced / fixed / last_affected, in order; "0" is the start).
// Any other range type, an unknown event or an irregular bound is
// Undetermined.
func EvalOSV(typ string, events []map[string]string, v string) Decision {
	if typ != "ECOSYSTEM" || !Regular(v) {
		return Undetermined
	}
	// Evaluate in version order (the OSV schema's algorithm); "0" first.
	sorted := append([]map[string]string(nil), events...)
	evVersion := func(ev map[string]string) string {
		for _, b := range ev {
			return b
		}
		return "0"
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := evVersion(sorted[i]), evVersion(sorted[j])
		if a == "0" || b == "0" {
			return a == "0" && b != "0"
		}
		return Compare(a, b) < 0
	})
	affected := false
	for _, ev := range sorted {
		for k, b := range ev {
			if b != "0" && !Regular(b) {
				return Undetermined
			}
			switch k {
			case "introduced":
				if b == "0" || Compare(v, b) >= 0 {
					affected = true
				}
			case "fixed":
				if Compare(v, b) >= 0 {
					affected = false
				}
			case "last_affected":
				if Compare(v, b) > 0 {
					affected = false
				}
			case "limit":
				if Compare(v, b) >= 0 {
					return Out
				}
			default:
				return Undetermined
			}
		}
	}
	if affected {
		return In
	}
	return Out
}
