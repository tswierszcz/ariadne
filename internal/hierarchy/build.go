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
		path  string
		gav   *input.GAV
		isApp bool
	}
	jobs := make([]job, 0, len(depJars)+1)
	jobs = append(jobs, job{path: appJarPath, isApp: true})
	for _, dj := range depJars {
		jobs = append(jobs, job{path: dj.Path, gav: dj.GAV})
	}

	type result struct {
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
				resCh <- result{path: j.path, classes: classes, err: err}
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

	classes := make(map[string]*ClassInfo)
	var warnings []string
	var firstErr error
	for r := range resCh {
		if r.err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("reading %s: %w", r.path, r.err)
			}
			continue
		}
		for _, ci := range r.classes {
			if existing, dup := classes[ci.FQCN]; dup {
				warnings = append(warnings, fmt.Sprintf(
					"class %s found in both %s and %s; keeping the first one seen",
					ci.FQCN, existing.SourceJar, ci.SourceJar))
				continue
			}
			classes[ci.FQCN] = ci
		}
	}
	if firstErr != nil {
		return nil, warnings, firstErr
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
