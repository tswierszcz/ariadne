package vulnmap

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"ariadne/internal/cves"
)

// promptVersion is part of every LLM cache key: bump it whenever the system
// prompt, the user template or the answer schema changes, so stale cached
// answers are not reused for a different prompt.
const promptVersion = 2 // 2: CVE:F5c fix_methods (both sides of the fix, release line)

const (
	maxDetailsChars = 8000
	maxDiffChars    = 60000
	maxFixMethods   = 80
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

// commitDiff is the kept (non-test .java) part of one fix commit. note,
// when set, follows the URL on its "# commit" line (CVE:F5c: a commit on
// another release line).
type commitDiff struct {
	url   string
	note  string
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

// keepForLLM is isNonTestJavaFile (see javasource.go) under the name this
// file's own callers originally used.
func keepForLLM(p string) bool {
	return isNonTestJavaFile(p)
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
		if c.note != "" {
			header = "# commit " + c.url + " (" + c.note + ")\n"
		}
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
	FixMethods string
}

// buildUserPrompt renders the user message for cve. commitURLs lists every
// fix commit the CVE links; commits holds the kept diffs of those that were
// fetched; fixes is the diff source's per-commit evidence (both sides of
// the fix, CVE:F5c), rendered as <fix_methods> when it names a method. It
// returns the message and what evidence it carries.
func buildUserPrompt(cve cves.CVE, commitURLs []string, commits []commitDiff, fixes []FixCommit) (prompt, evidence string, err error) {
	d := promptData{
		FixMethods: renderFixMethods(fixes),
		ID:         cve.ID,
		Aliases:    orNone(strings.Join(cve.Aliases, ", ")),
		Summary:    orNone(cve.Summary),
		Details:    orNone(truncateChars(cve.Details, maxDetailsChars)),
		Packages:   orNone(joinPackages(cve.Affected)),
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

// publicCommitURL names a commit by its GitHub URL, whatever base URL a
// test points the resolver at.
func publicCommitURL(repo, sha string) string {
	return "https://github.com/" + repo + "/commit/" + sha
}

// lineNote describes a commit's release line for the prompt.
func lineNote(c FixCommit) string {
	var n string
	switch c.ReleaseLine {
	case LineDescends:
		n = "on the analyzed release's line (tag " + c.ReleaseTag + ")"
	case LineInRelease:
		n = "already contained in the analyzed release (tag " + c.ReleaseTag + ")"
	case LineOther:
		n = "on another release line than the analyzed release (tag " + c.ReleaseTag + ")"
	}
	if c.Excluded != "" {
		if n != "" {
			n += "; "
		}
		n += "not the fix of the analyzed release"
	}
	return n
}

var fixKindText = map[string]string{
	FixMethodChanged: "changed by the fix (declared before it)",
	FixMethodRemoved: "removed by the fix (declared before it only)",
	FixMethodAdded:   "added by the fix (not in the vulnerable release)",
}

// renderFixMethods lists, per readable commit, the methods the fix
// touched as read from the source before and after it, with whether each
// exists in the analyzed classes. "" when no commit names a method.
func renderFixMethods(fixes []FixCommit) string {
	var b strings.Builder
	n, omitted := 0, 0
	for _, c := range fixes {
		if c.Error != "" {
			continue
		}
		var lines []string
		for _, f := range c.Files {
			for _, m := range f.Methods {
				if n >= maxFixMethods {
					omitted++
					continue
				}
				n++
				sel := m.Selector
				if !strings.Contains(sel, "(") {
					sel += "(" + strings.Join(m.SourceParams, ", ") + ")"
				}
				kind := fixKindText[m.Kind]
				if c.Base == BaseNewSideOnly || f.NewSideOnly {
					kind = "touched by the fix (pre-fix source unavailable)"
				}
				line := "- " + sel + ": " + kind
				switch m.Bytecode {
				case BytecodeExact, BytecodeAmbiguous, BytecodeParamMismatch:
					line += "; present in the analyzed classes"
				case BytecodeMethodMissing, BytecodeClassMissing:
					line += "; not in the analyzed classes"
				}
				lines = append(lines, line)
			}
		}
		if len(lines) == 0 {
			continue
		}
		b.WriteString("# commit " + publicCommitURL(c.Repo, c.SHA))
		if note := lineNote(c); note != "" {
			b.WriteString(" (" + note + ")")
		}
		b.WriteString("\n" + strings.Join(lines, "\n") + "\n")
	}
	if omitted > 0 {
		fmt.Fprintf(&b, "# [truncated: %d more methods omitted]\n", omitted)
	}
	return strings.TrimRight(b.String(), "\n")
}
