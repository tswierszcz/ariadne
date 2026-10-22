package hierarchy

import (
	"fmt"
	"sort"
	"sync"

	"ariadne/internal/input"
)

const readWorkers = 8

// Build reads every class from the application (appJarPath: a plain jar
// or a jar packed from a classes directory, per Phase 1) and from every
// dependency jar, in parallel, and assembles the resulting Hierarchy.
//
// A class that's defined in more than one jar (a duplicate on the
// classpath — not unusual with overlapping shaded dependencies) is kept
// from whichever jar was read first and reported as a warning rather than
// an error, since ariadne can still analyze the rest of the graph.
func Build(appJarPath string, depJars []input.DepJar) (*Hierarchy, []string, error) {
	type job struct {
		idx   int
		path  string
		gav   *input.GAV
		isApp bool
	}
	jobs := make([]job, 0, len(depJars)+1)
	jobs = append(jobs, job{idx: 0, path: appJarPath, isApp: true})
	for i, dj := range depJars {
		jobs = append(jobs, job{idx: i + 1, path: dj.Path, gav: dj.GAV})
	}

	type result struct {
		idx     int
		path    string
		classes []*ClassInfo
		err     error
	}
	jobCh := make(chan job)
	resCh := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < readWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				classes, err := readJarClasses(j.path, j.gav, j.isApp)
				resCh <- result{idx: j.idx, path: j.path, classes: classes, err: err}
			}
		}()
	}
	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
	}()
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// Results are collected by job index, not in whatever order workers
	// happen to finish: JVM_GRAPH G1's own correctness requirement is that
	// which duplicate class "wins" depends only on the explicit input
	// order (app, then dependency jars in the order the CLI/ResolveDepJars
	// gave them) — never on worker scheduling, jar size, or disk speed,
	// none of which are supposed to change the analysis's own answer.
	results := make([]result, len(jobs))
	var firstErr error
	for r := range resCh {
		if r.err != nil && firstErr == nil {
			firstErr = fmt.Errorf("reading %s: %w", r.path, r.err)
		}
		results[r.idx] = r
	}
	if firstErr != nil {
		return nil, nil, firstErr
	}

	classes := make(map[string]*ClassInfo)
	var warnings []string
	var duplicates []Duplicate
	for _, r := range results {
		for _, ci := range r.classes {
			existing, dup := classes[ci.FQCN]
			if !dup {
				classes[ci.FQCN] = ci
				continue
			}
			warnings = append(warnings, fmt.Sprintf(
				"class %s found in both %s and %s; keeping the first one seen (explicit input order, not read order)",
				ci.FQCN, existing.SourceJar, ci.SourceJar))
			duplicates = append(duplicates, Duplicate{
				FQCN:     ci.FQCN,
				Kept:     sourceRefOf(existing),
				Shadowed: sourceRefOf(ci),
			})
		}
	}

	sorted := make([]string, 0, len(classes))
	for fqcn := range classes {
		sorted = append(sorted, fqcn)
	}
	sort.Strings(sorted)

	return &Hierarchy{
		classes:       classes,
		direct:        buildDirectSubtypes(classes, sorted),
		subtypesCache: make(map[string][]string),
		sorted:        sorted,
		duplicates:    duplicates,
	}, warnings, nil
}

// buildDirectSubtypes inverts the superclass/interfaces edges every class
// already carries, so Subtypes can walk "who extends/implements me"
// instead of "what do I extend/implement". It iterates sortedFQCNs
// (rather than ranging over classes directly) purely for determinism: the
// order each supertype's direct-subtype slice gets built in otherwise
// depends on Go's randomized map iteration order, which would make
// Subtypes' result order — and from there, which of several equally
// short CHA-expanded paths internal/reach ends up reporting — differ
// between two runs of the exact same analysis.
func buildDirectSubtypes(classes map[string]*ClassInfo, sortedFQCNs []string) map[string][]string {
	direct := make(map[string][]string)
	for _, fqcn := range sortedFQCNs {
		ci := classes[fqcn]
		if ci.SuperClass != "" {
			direct[ci.SuperClass] = append(direct[ci.SuperClass], fqcn)
		}
		for _, iface := range ci.Interfaces {
			direct[iface] = append(direct[iface], fqcn)
		}
	}
	return direct
}
