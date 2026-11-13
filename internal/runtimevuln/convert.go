package runtimevuln

import (
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Conversion of hand-saved OpenJDK advisory pages into catalog entries
// (cmd/openjdk-catalog). The output is a draft: the maintainer checks it
// against the page before committing it.

var (
	tableRe   = regexp.MustCompile(`(?is)<table\b.*?</table>`)
	rowRe     = regexp.MustCompile(`(?is)<tr\b.*?</tr>`)
	cellRe    = regexp.MustCompile(`(?is)<t([hd])\b[^>]*>(.*?)</t[hd]>`)
	tagRe     = regexp.MustCompile(`(?s)<[^>]*>`)
	spaceRe   = regexp.MustCompile(`\s+`)
	affectRe  = regexp.MustCompile(`(?i)affected versions are\s+(.+?)\.(?:\s+[A-Z]|\s*$)`)
	versionRe = regexp.MustCompile(`\b(\d+u\d+|\d+(?:\.\d+){0,3})\b`)
	cvssRe    = regexp.MustCompile(`^\d+(\.\d+)?$`)
)

func text(s string) string {
	return strings.TrimSpace(spaceRe.ReplaceAllString(html.UnescapeString(tagRe.ReplaceAllString(s, " ")), " "))
}

// ParseAdvisoryHTML converts one saved advisory page. date is the
// advisory's date (YYYY-MM-DD), taken from the saved file's name.
func ParseAdvisoryHTML(date, page string) (Advisory, error) {
	a := Advisory{Date: date, URL: "https://openjdk.org/groups/vulnerability/advisories/" + date, LastAffected: map[string]string{}}
	body := text(page)
	m := affectRe.FindStringSubmatch(body)
	if m == nil {
		return a, fmt.Errorf("%s: no \"affected versions are …\" sentence", date)
	}
	list := m[1]
	a.EarlierAffected = strings.Contains(strings.ToLower(list), "earlier")
	for _, v := range versionRe.FindAllString(list, -1) {
		rel, ok := ParseUpstream(v)
		if !ok {
			return a, fmt.Errorf("%s: affected version %q", date, v)
		}
		a.LastAffected[strconv.Itoa(rel.Feature)] = v
	}
	if len(a.LastAffected) == 0 {
		return a, fmt.Errorf("%s: no versions in %q", date, list)
	}
	// The OpenJDK risk matrix is the first table whose header names the
	// CVE column; a page may also carry an OpenJFX matrix after it.
	for _, t := range tableRe.FindAllString(page, -1) {
		rows := rowRe.FindAllString(t, -1)
		header := -1
		var lines []int
		for i, r := range rows {
			cells := cellRe.FindAllStringSubmatch(r, -1)
			if len(cells) == 0 || !strings.Contains(strings.ToUpper(text(r)), "CVE ID") {
				continue
			}
			// A header can span two rows ("Affects ..." over the line
			// numbers): the header is the first one naming lines.
			var nums []int
			for _, c := range cells {
				if n, err := strconv.Atoi(text(c[2])); err == nil {
					nums = append(nums, n)
				}
			}
			if len(nums) == 0 {
				continue
			}
			header, lines = i, nums
			break
		}
		if header < 0 || len(lines) == 0 {
			continue
		}
		for _, r := range rows[header+1:] {
			cells := cellRe.FindAllStringSubmatch(r, -1)
			if len(cells) < 2+len(lines) {
				continue
			}
			id := text(cells[0][2])
			if !cveRe.MatchString(id) {
				continue
			}
			cve := CVE{ID: id, Component: strings.ReplaceAll(text(cells[1][2]), " ", "")}
			if len(cells) > 2 {
				if s := text(cells[2][2]); cvssRe.MatchString(s) {
					cve.CVSS, _ = strconv.ParseFloat(s, 64)
				}
			}
			// Line columns are the last len(lines) cells.
			marks := cells[len(cells)-len(lines):]
			for i, c := range marks {
				if text(c[2]) != "" {
					cve.Lines = append(cve.Lines, lines[i])
				}
			}
			for _, l := range cve.Lines {
				if _, ok := a.LastAffected[strconv.Itoa(l)]; !ok {
					return a, fmt.Errorf("%s: %s marks line %d, which the affected-versions sentence does not name", date, id, l)
				}
			}
			a.CVEs = append(a.CVEs, cve)
		}
		break
	}
	if len(a.CVEs) == 0 {
		return a, fmt.Errorf("%s: no OpenJDK risk matrix rows found", date)
	}
	sort.Slice(a.CVEs, func(i, j int) bool { return a.CVEs[i].ID < a.CVEs[j].ID })
	return a, nil
}
