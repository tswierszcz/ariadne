// Package runtimeevents imports a runtime trace (RUNTIME_NATIVE N0/N1,
// docs/RUNTIME_TRACE.MD: format ariadne-runtime-trace-1) and correlates it
// with the analyzed input: which recorded classes are the analyzed classes
// (by content hash, never by name alone), which recorded methods are graph
// nodes, and whether each measured caller → callee relation is an edge of
// the static graph. It changes no edge and no verdict: absence of an event
// proves nothing, and execution is not exploitation.
package runtimeevents

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Format is the only accepted trace format.
const Format = "ariadne-runtime-trace-1"

// ImportVersion versions the importer's correlation rules (snapshots record it).
const ImportVersion = "n1-1"

// Limits on the input (docs/RUNTIME_TRACE.MD).
var (
	MaxBytes   int64 = 1 << 30
	MaxLine          = 1 << 20
	MaxRecords       = 5_000_000
	MaxString        = 4096
)

// Header is the trace's first record.
type Header struct {
	Record   string `json:"record"`
	Format   string `json:"format"`
	RunID    string `json:"run_id"`
	Scenario string `json:"scenario"`
	Started  string `json:"started"`
	Ended    string `json:"ended"`
	Runtime  struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Vendor  string `json:"vendor"`
	} `json:"runtime"`
	Recorder struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Source  string `json:"source"`
	} `json:"recorder"`
	Scope struct {
		EventKinds    []string `json:"event_kinds"`
		Classes       string   `json:"classes"`
		MethodFilters []string `json:"method_filters"`
	} `json:"scope"`
	Sampling struct {
		Kind       string `json:"kind"`
		IntervalMS int    `json:"interval_ms"`
	} `json:"sampling"`
	Completeness struct {
		Complete        bool     `json:"complete"`
		LostEvents      int      `json:"lost_events"`
		TruncatedStacks int      `json:"truncated_stacks"`
		Reasons         []string `json:"reasons"`
	} `json:"completeness"`
	DataPolicy string `json:"data_policy"`
}

// Artifact, Loader, Class, Method are the trace's identities.
type Artifact struct{ ID, Name, SHA256 string }
type Loader struct{ ID, Name, Type string }
type Class struct {
	ID, Name, Loader, Artifact, SHA256, SourceBasis string
}
type Method struct{ ID, Class, Name, Desc string }

// Loaded, Executed, Call and Handoff are the evidence records; their IDs
// are the evidence ids results refer to.
type Loaded struct{ ID, Class string }
type Executed struct {
	ID, Method, Basis string
	Count             int
}
type Call struct {
	ID, Caller, Callee, Basis, Source string
	Count                             int
	BCI, Line                         int // -1 when not recorded
}
type Handoff struct{ ID, From, To, Mechanism, Basis string }

// Trace is a parsed, validated trace.
type Trace struct {
	Header    Header
	SHA256    string
	Artifacts map[string]Artifact
	Loaders   map[string]Loader
	Classes   map[string]Class
	Methods   map[string]Method
	Loaded    []Loaded
	Executed  []Executed
	Calls     []Call
	Handoffs  []Handoff

	defined map[string]bool // loader + "/" + name
}

type record struct {
	Record string `json:"record"`
	ID     string `json:"id"`
	// artifact / loader / class / method
	Name        string `json:"name"`
	SHA256      string `json:"sha256"`
	Type        string `json:"type"`
	Loader      string `json:"loader"`
	Artifact    string `json:"artifact"`
	SourceBasis string `json:"source_basis"`
	Class       string `json:"class"`
	Desc        string `json:"desc"`
	// evidence
	Method    string `json:"method"`
	Basis     string `json:"basis"`
	Count     *int   `json:"count"`
	Caller    string `json:"caller"`
	Callee    string `json:"callee"`
	Source    string `json:"source"`
	BCI       *int   `json:"bci"`
	Line      *int   `json:"line"`
	From      string `json:"from"`
	To        string `json:"to"`
	Mechanism string `json:"mechanism"`
}

var allowedFields = map[string][]string{
	"artifact": {"record", "id", "name", "sha256"},
	"loader":   {"record", "id", "name", "type"},
	"class":    {"record", "id", "name", "loader", "artifact", "sha256", "source_basis"},
	"method":   {"record", "id", "class", "name", "desc"},
	"loaded":   {"record", "id", "class"},
	"executed": {"record", "id", "method", "basis", "count"},
	"call":     {"record", "id", "caller", "callee", "basis", "source", "count", "bci", "line"},
	"handoff":  {"record", "id", "from", "to", "mechanism", "basis"},
}

// Parse reads and validates a trace; sha is the file's SHA-256 (the
// caller hashes the bytes it read). Any violation rejects the whole file.
func Parse(r io.Reader, sha string) (*Trace, error) {
	t := &Trace{SHA256: sha, Artifacts: map[string]Artifact{}, Loaders: map[string]Loader{}, Classes: map[string]Class{}, Methods: map[string]Method{}, defined: map[string]bool{}}
	sc := bufio.NewScanner(io.LimitReader(r, MaxBytes+1))
	sc.Buffer(make([]byte, 64*1024), MaxLine+1)
	line, read := 0, int64(0)
	evidence := map[string]bool{}
	for sc.Scan() {
		line++
		b := sc.Bytes()
		read += int64(len(b)) + 1
		if read > MaxBytes {
			return nil, fmt.Errorf("trace larger than %d bytes", MaxBytes)
		}
		if line > MaxRecords+1 {
			return nil, fmt.Errorf("trace has more than %d records", MaxRecords)
		}
		if len(bytes.TrimSpace(b)) == 0 {
			return nil, fmt.Errorf("line %d: empty line", line)
		}
		if line == 1 {
			if err := strictDecode(b, &t.Header); err != nil {
				return nil, fmt.Errorf("line 1 (header): %w", err)
			}
			if err := t.Header.validate(); err != nil {
				return nil, fmt.Errorf("line 1 (header): %w", err)
			}
			continue
		}
		var rec record
		if err := json.Unmarshal(b, &rec); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if err := checkFields(b, rec.Record); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if err := t.add(rec, evidence); err != nil {
			return nil, fmt.Errorf("line %d (%s %q): %w", line, rec.Record, rec.ID, err)
		}
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("line %d: longer than %d bytes", line+1, MaxLine)
		}
		return nil, err
	}
	if line == 0 {
		return nil, errors.New("empty trace: no header")
	}
	return t, nil
}

func strictDecode(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func checkFields(b []byte, kind string) error {
	allowed, ok := allowedFields[kind]
	if !ok {
		return fmt.Errorf("unknown record kind %q", kind)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		found := false
		for _, a := range allowed {
			if a == k {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown field %q in a %s record", k, kind)
		}
		if len(v) > MaxString+2 && v[0] == '"' {
			return fmt.Errorf("field %q longer than %d bytes", k, MaxString)
		}
	}
	return nil
}

func oneOf(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

func (h Header) validate() error {
	if h.Record != "header" {
		return fmt.Errorf("first record is %q, want \"header\"", h.Record)
	}
	if h.Format != Format {
		return fmt.Errorf("format %q, want %q", h.Format, Format)
	}
	if h.RunID == "" || h.Recorder.Name == "" || h.Recorder.Version == "" {
		return errors.New("run_id, recorder.name and recorder.version are required")
	}
	for _, k := range h.Scope.EventKinds {
		if !oneOf(k, "loaded", "executed", "call", "handoff") {
			return fmt.Errorf("unknown scope.event_kinds value %q", k)
		}
	}
	if !oneOf(h.Scope.Classes, "artifacts", "all") {
		return fmt.Errorf("scope.classes %q, want \"artifacts\" or \"all\"", h.Scope.Classes)
	}
	if !oneOf(h.Sampling.Kind, "none", "periodic") {
		return fmt.Errorf("sampling.kind %q, want \"none\" or \"periodic\"", h.Sampling.Kind)
	}
	if h.Sampling.Kind == "periodic" && h.Completeness.Complete {
		return errors.New("a periodically sampled trace cannot be complete")
	}
	if h.Completeness.LostEvents < 0 || h.Completeness.TruncatedStacks < 0 {
		return errors.New("negative completeness counts")
	}
	if h.DataPolicy != "identifiers-only" {
		return fmt.Errorf("data_policy %q, want \"identifiers-only\"", h.DataPolicy)
	}
	return nil
}

// recorded reports whether the header declares the event kind.
func (h Header) recorded(kind string) bool {
	for _, k := range h.Scope.EventKinds {
		if k == kind {
			return true
		}
	}
	return false
}

func (t *Trace) add(r record, evidence map[string]bool) error {
	if r.ID == "" {
		return errors.New("missing id")
	}
	dup := func(exists bool) error {
		if exists {
			return errors.New("duplicate id")
		}
		return nil
	}
	ev := func(kind string) error {
		if !t.Header.recorded(kind) {
			return fmt.Errorf("a %s record, but the header's scope.event_kinds does not declare %q", kind, kind)
		}
		if evidence[r.ID] {
			return errors.New("duplicate evidence id")
		}
		evidence[r.ID] = true
		return nil
	}
	count := func() (int, error) {
		if r.Count == nil || *r.Count < 1 {
			return 0, errors.New("count must be at least 1")
		}
		return *r.Count, nil
	}
	method := func(id string) error {
		if _, ok := t.Methods[id]; !ok {
			return fmt.Errorf("unknown method %q", id)
		}
		return nil
	}
	switch r.Record {
	case "artifact":
		if _, ok := t.Artifacts[r.ID]; ok {
			return dup(true)
		}
		if !isSHA256(r.SHA256) {
			return errors.New("sha256 must be 64 lowercase hex digits")
		}
		t.Artifacts[r.ID] = Artifact{ID: r.ID, Name: r.Name, SHA256: r.SHA256}
	case "loader":
		if _, ok := t.Loaders[r.ID]; ok {
			return dup(true)
		}
		t.Loaders[r.ID] = Loader{ID: r.ID, Name: r.Name, Type: r.Type}
	case "class":
		if _, ok := t.Classes[r.ID]; ok {
			return dup(true)
		}
		if r.Name == "" || strings.ContainsAny(r.Name, "/;[") {
			return fmt.Errorf("class name %q is not a dotted binary name", r.Name)
		}
		if _, ok := t.Loaders[r.Loader]; !ok {
			return fmt.Errorf("unknown loader %q", r.Loader)
		}
		if r.Artifact != "" {
			if _, ok := t.Artifacts[r.Artifact]; !ok {
				return fmt.Errorf("unknown artifact %q", r.Artifact)
			}
		}
		if r.SHA256 != "" && !isSHA256(r.SHA256) {
			return errors.New("sha256 must be 64 lowercase hex digits")
		}
		if !oneOf(r.SourceBasis, "observed", "classpath-lookup") {
			return fmt.Errorf("source_basis %q, want \"observed\" or \"classpath-lookup\"", r.SourceBasis)
		}
		if t.defined[r.Loader+"/"+r.Name] {
			return fmt.Errorf("class %s is already defined by loader %s", r.Name, r.Loader)
		}
		t.defined[r.Loader+"/"+r.Name] = true
		t.Classes[r.ID] = Class{ID: r.ID, Name: r.Name, Loader: r.Loader, Artifact: r.Artifact, SHA256: r.SHA256, SourceBasis: r.SourceBasis}
	case "method":
		if _, ok := t.Methods[r.ID]; ok {
			return dup(true)
		}
		if _, ok := t.Classes[r.Class]; !ok {
			return fmt.Errorf("unknown class %q", r.Class)
		}
		if r.Name == "" || !strings.HasPrefix(r.Desc, "(") || !strings.Contains(r.Desc, ")") {
			return fmt.Errorf("method %q has no JVM descriptor (%q)", r.Name, r.Desc)
		}
		t.Methods[r.ID] = Method{ID: r.ID, Class: r.Class, Name: r.Name, Desc: r.Desc}
	case "loaded":
		if err := ev("loaded"); err != nil {
			return err
		}
		if _, ok := t.Classes[r.Class]; !ok {
			return fmt.Errorf("unknown class %q", r.Class)
		}
		t.Loaded = append(t.Loaded, Loaded{ID: r.ID, Class: r.Class})
	case "executed":
		if err := ev("executed"); err != nil {
			return err
		}
		if err := method(r.Method); err != nil {
			return err
		}
		if !oneOf(r.Basis, "traced", "sampled") {
			return fmt.Errorf("basis %q, want \"traced\" or \"sampled\"", r.Basis)
		}
		n, err := count()
		if err != nil {
			return err
		}
		t.Executed = append(t.Executed, Executed{ID: r.ID, Method: r.Method, Basis: r.Basis, Count: n})
	case "call":
		if err := ev("call"); err != nil {
			return err
		}
		if err := method(r.Caller); err != nil {
			return err
		}
		if err := method(r.Callee); err != nil {
			return err
		}
		if !oneOf(r.Basis, "stack-adjacent", "stack-through-runtime") {
			return fmt.Errorf("basis %q, want \"stack-adjacent\" or \"stack-through-runtime\"", r.Basis)
		}
		if !oneOf(r.Source, "traced", "sampled") {
			return fmt.Errorf("source %q, want \"traced\" or \"sampled\"", r.Source)
		}
		n, err := count()
		if err != nil {
			return err
		}
		c := Call{ID: r.ID, Caller: r.Caller, Callee: r.Callee, Basis: r.Basis, Source: r.Source, Count: n, BCI: -1, Line: -1}
		if r.BCI != nil {
			c.BCI = *r.BCI
		}
		if r.Line != nil {
			c.Line = *r.Line
		}
		t.Calls = append(t.Calls, c)
	case "handoff":
		if err := ev("handoff"); err != nil {
			return err
		}
		if err := method(r.From); err != nil {
			return err
		}
		if err := method(r.To); err != nil {
			return err
		}
		if r.Mechanism == "" || r.Basis == "" {
			return errors.New("mechanism and basis are required")
		}
		t.Handoffs = append(t.Handoffs, Handoff{ID: r.ID, From: r.From, To: r.To, Mechanism: r.Mechanism, Basis: r.Basis})
	}
	return nil
}

func isSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// sortedIDs returns map keys in a stable order.
func sortedIDs[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
