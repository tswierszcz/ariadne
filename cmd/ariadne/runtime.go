package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"ariadne/internal/component"
	"ariadne/internal/hierarchy"
	"ariadne/internal/reach"
	"ariadne/internal/report"
	"ariadne/internal/runtimeevents"
)

// loadRuntimeTrace reads and validates -runtime-trace before the analysis,
// so a bad file fails early (RUNTIME_NATIVE N1).
func loadRuntimeTrace(path string) (*runtimeevents.Trace, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace: %w", err)
	}
	if st.Size() > runtimeevents.MaxBytes {
		return nil, fmt.Errorf("-runtime-trace %s: %d bytes, more than the %d-byte limit", path, st.Size(), runtimeevents.MaxBytes)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace: %w", err)
	}
	sum := sha256.Sum256(b)
	t, err := runtimeevents.Parse(bytes.NewReader(b), hex.EncodeToString(sum[:]))
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace %s: %w", path, err)
	}
	return t, nil
}

// classLookup answers runtimeevents.Inputs.Lookup from the analyzed
// classes: the SHA-256 of the class file actually read (its jar entry) and
// of the jar it was read from. Jars are opened once.
func classLookup(h *hierarchy.Hierarchy) (func(string) (string, string, bool), func()) {
	jars := map[string]*zip.ReadCloser{}
	jarSHA := map[string]string{}
	classSHA := map[string]string{}
	closeAll := func() {
		for _, z := range jars {
			z.Close()
		}
	}
	return func(fqcn string) (string, string, bool) {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return "", "", false
		}
		if _, done := jarSHA[ci.SourceJar]; !done {
			b, err := os.ReadFile(ci.SourceJar)
			if err == nil {
				s := sha256.Sum256(b)
				jarSHA[ci.SourceJar] = hex.EncodeToString(s[:])
			} else {
				jarSHA[ci.SourceJar] = ""
			}
		}
		if cs, done := classSHA[fqcn]; done {
			return cs, jarSHA[ci.SourceJar], true
		}
		cs := ""
		z, ok := jars[ci.SourceJar]
		if !ok {
			if zr, err := zip.OpenReader(ci.SourceJar); err == nil {
				z = zr
				jars[ci.SourceJar] = zr
			}
		}
		if z != nil {
			for _, f := range z.File {
				if f.Name != ci.SourceEntry {
					continue
				}
				if rc, err := f.Open(); err == nil {
					hsh := sha256.New()
					if _, err := io.Copy(hsh, rc); err == nil {
						cs = hex.EncodeToString(hsh.Sum(nil))
					}
					rc.Close()
				}
				break
			}
		}
		classSHA[fqcn] = cs
		return cs, jarSHA[ci.SourceJar], true
	}, closeAll
}

// runtimeObserved: per advisory and component instance, the mapped
// vulnerable methods seen executing in the trace.
func runtimeObserved(rt runtimeevents.Result, results []reach.Result, h *hierarchy.Hierarchy, inv *component.Inventory) map[string]map[string][]runtimeevents.MethodObservation {
	out := map[string]map[string][]runtimeevents.MethodObservation{}
	for _, r := range results {
		for _, o := range rt.ObservedVulnerable(r.Methods) {
			owner, _, _ := strings.Cut(o.Method, "#")
			inst := ""
			if ci, ok := h.Lookup(owner); ok && inv != nil {
				inst = inv.InstanceForClass(ci)
			}
			if out[r.CVE] == nil {
				out[r.CVE] = map[string][]runtimeevents.MethodObservation{}
			}
			out[r.CVE][inst] = append(out[r.CVE][inst], o)
		}
	}
	return out
}

func runtimeFindings(obs map[string]map[string][]runtimeevents.MethodObservation, results []reach.Result) []report.RuntimeFinding {
	var out []report.RuntimeFinding
	for _, r := range results {
		var all []runtimeevents.MethodObservation
		for _, os := range obs[r.CVE] {
			all = append(all, os...)
		}
		if len(all) == 0 {
			continue
		}
		sort.Slice(all, func(i, j int) bool { return all[i].Evidence < all[j].Evidence })
		out = append(out, report.RuntimeFinding{Advisory: r.CVE, Status: string(r.Status), Observed: all})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Advisory < out[j].Advisory })
	return out
}
