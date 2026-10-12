package vulnmap

import (
	"bytes"
	_ "embed"
	"fmt"
	"path"
	"strings"
	"text/template"

	"ariadne/internal/cves"
)

// promptVersion is part of every LLM cache key: bump it whenever the system
// prompt, the user template or the answer schema changes, so stale cached
// answers are not reused for a different prompt.
const promptVersion = 1

const (
	maxDetailsChars = 8000
	maxDiffChars    = 60000
)

//go:embed llm_system_prompt.txt
var llmSystemPromptRaw string

//go:embed llm_user_prompt.tmpl
var llmUserPromptTmpl string

var userPromptTemplate = template.Must(template.New("user").Parse(llmUserPromptTmpl))

// llmSystemPrompt is the system message, without the file's trailing newline.
func llmSystemPrompt() string { return strings.TrimRight(llmSystemPromptRaw, "\n") }

// correctionMessage is the user turn appended on the single corrective retry.
func correctionMessage(problem string) string {
	return fmt.Sprintf("Your previous reply could not be used: %s. Reply again with only the JSON object, in exactly the shape given in the system message.", problem)
}

// What the model was shown for one CVE.
const (
	EvidenceDiff     = "diff"     // the fix commit's non-test .java changes
	EvidenceAdvisory = "advisory" // the advisory text only
)

// fileDiff is one "diff --git" section of a fix commit's diff.
type fileDiff struct {
	path string
	text string
}

// commitDiff is the kept (non-test .java) part of one fix commit.
type commitDiff struct {
	url   string
	files []fileDiff
}

// splitDiffFiles cuts a unified diff into its per-file "diff --git" sections.
func splitDiffFiles(diff string) []fileDiff {
	var out []fileDiff
	var cur *fileDiff
	for _, line := range strings.SplitAfter(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			out = append(out, fileDiff{path: diffGitPath(line)})
			cur = &out[len(out)-1]
		}
		if cur != nil {
			cur.text += line
		}
	}
	return out
}

// diffGitPath extracts the new-side path from "diff --git a/X b/Y".
func diffGitPath(header string) string {
	header = strings.TrimSpace(header)
	if i := strings.LastIndex(header, " b/"); i >= 0 {
		return header[i+len(" b/"):]
	}
	return ""
}

// isTestPath reports whether a file path is test code: any path segment
// named test or tests, or a file named *Test.java / *Tests.java.
func isTestPath(p string) bool {
	for _, seg := range strings.Split(path.Dir(p), "/") {
		if seg == "test" || seg == "tests" {
			return true
		}
	}
	base := path.Base(p)
	return strings.HasSuffix(base, "Test.java") || strings.HasSuffix(base, "Tests.java")
}

// keepForLLM keeps non-test .java files only. Unlike source 2 it does not
// require src/main/: that is exactly what hides Tomcat's java/org/apache/...
// layout from source 2, and the model can judge those files fine.
func keepForLLM(p string) bool {
	return strings.HasSuffix(p, ".java") && !isTestPath(p)
}

// assembleDiff concatenates the kept files of every commit, in order, under
// a "# commit <url>" line each. When the total would exceed limit it stops
// at a file boundary and ends with a "# [truncated: ...]" line naming the
// omitted files. A first file that alone exceeds limit is cut at a line
// boundary rather than dropped, so the model always sees something.
func assembleDiff(commits []commitDiff, limit int) (text string, truncated bool) {
	var b strings.Builder
	var omitted []string

	for _, c := range commits {
		if len(c.files) == 0 {
			continue
		}
		header := "# commit " + c.url + "\n"
		headerWritten := false
		for _, f := range c.files {
			if truncated {
				omitted = append(omitted, f.path)
				continue
			}
			need := len(f.text)
			if !headerWritten {
				need += len(header)
			}
			if b.Len()+need > limit {
				if b.Len() == 0 {
					b.WriteString(header)
					headerWritten = true
					b.WriteString(cutAtLine(f.text, limit-len(header)))
				} else {
					omitted = append(omitted, f.path)
				}
				truncated = true
				continue
			}
			if !headerWritten {
				b.WriteString(header)
				headerWritten = true
			}
			b.WriteString(f.text)
		}
	}
	if truncated {
		if len(omitted) > 0 {
			fmt.Fprintf(&b, "# [truncated: %d more files omitted: %s]\n", len(omitted), strings.Join(omitted, ", "))
		} else {
			b.WriteString("# [truncated: the last file was cut to fit]\n")
		}
	}
	return strings.TrimRight(b.String(), "\n"), truncated
}

func cutAtLine(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	cut := s[:n]
	if i := strings.LastIndexByte(cut, '\n'); i >= 0 {
		cut = cut[:i+1]
	}
	return cut
}

type promptData struct {
	Variant    string // "A" diff, "B" commits without Java, "C" no commit
	ID         string
	Aliases    string
	Packages   string
	Summary    string
	Details    string
	Diff       string
	Truncated  bool
	CommitURLs string
}

// buildUserPrompt renders the user message for cve. commitURLs lists every
// fix commit the CVE links; commits holds the kept diffs of those that were
// fetched. It returns the message and what evidence it carries.
func buildUserPrompt(cve cves.CVE, commitURLs []string, commits []commitDiff) (prompt, evidence string, err error) {
	d := promptData{
		ID:       cve.ID,
		Aliases:  orNone(strings.Join(cve.Aliases, ", ")),
		Summary:  orNone(cve.Summary),
		Details:  orNone(truncateChars(cve.Details, maxDetailsChars)),
		Packages: orNone(joinPackages(cve.Affected)),
	}

	diff, truncated := assembleDiff(commits, maxDiffChars)
	switch {
	case diff != "":
		d.Variant, d.Diff, d.Truncated, evidence = "A", diff, truncated, EvidenceDiff
	case len(commitURLs) > 0:
		d.Variant, d.CommitURLs, evidence = "B", strings.Join(commitURLs, ", "), EvidenceAdvisory
	default:
		d.Variant, evidence = "C", EvidenceAdvisory
	}

	var buf bytes.Buffer
	if err := userPromptTemplate.Execute(&buf, d); err != nil {
		return "", "", fmt.Errorf("rendering LLM user prompt: %w", err)
	}
	return strings.TrimRight(buf.String(), "\n"), evidence, nil
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}

func joinPackages(aff []cves.Affected) string {
	parts := make([]string, len(aff))
	for i, a := range aff {
		parts[i] = a.Package + " " + a.Version
	}
	return strings.Join(parts, "; ")
}

func truncateChars(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "\n[truncated]"
}
