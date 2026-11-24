package research

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Resource limits a reader applies before reading (CONTRACT.md §4).
const (
	maxManifestBytes = 1 << 20
	maxIdentityBytes = 16 << 20
	maxDatasetBytes  = 4 << 30
	maxDatasets      = 64
	maxLineBytes     = 64 << 20
)

// Error is a failure with a contract error code (CONTRACT.md §11).
type Error struct {
	Code     string
	Message  string
	Location string // JSON Pointer or CLI option; "" when none
	Reason   string
}

func (e *Error) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("%s: %s (%s: %s)", e.Code, e.Message, e.Location, e.Reason)
	}
	return e.Code + ": " + e.Message
}

// ExitCode is the CLI exit status for the error (CONTRACT.md §3, §11).
func (e *Error) ExitCode() int {
	switch e.Code {
	case "INVALID_REQUEST", "UNSUPPORTED_SCHEMA", "UNSUPPORTED_OPERATION", "UNSUPPORTED_FILTER",
		"ANALYSIS_MISMATCH", "INVALID_CURSOR", "ROOTS_OUTSIDE_INDEX":
		return 2
	}
	return 1
}

func corrupt(format string, a ...any) *Error {
	return &Error{Code: "BUNDLE_CORRUPT", Message: fmt.Sprintf(format, a...)}
}

// datasetFiles are the datasets a producer writes, one per record type.
var datasetFiles = []struct{ path, recordType string }{
	{"components.jsonl", TypeComponent},
	{"classes.jsonl", TypeClass},
	{"nodes.jsonl", TypeNode},
	{"edges.jsonl", TypeEdge},
	{"roots.jsonl", TypeRoot},
	{"gaps.jsonl", TypeGap},
	{"calls.jsonl", TypeCall},
}

// Publish writes a complete bundle to dir, which must not exist: the
// bundle is written to a sibling ".<name>.partial-*" directory, hashed,
// and renamed into place, under a "<dir>.lock" file that makes a
// concurrent publisher fail (CONTRACT.md §3). ident's Datasets are
// filled in here. It returns the analysis id.
func Publish(dir string, ident Identity, recs Records, createdAt time.Time) (string, error) {
	dir = filepath.Clean(dir)
	lock := dir + ".lock"
	lf, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", fmt.Errorf("%s exists: another publication of %s is in progress (or was interrupted; remove the lock file)", lock, dir)
		}
		return "", err
	}
	lf.Close()
	defer os.Remove(lock)
	if _, err := os.Lstat(dir); err == nil {
		return "", fmt.Errorf("%s already exists: a bundle is never overwritten", dir)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	var rnd [6]byte
	rand.Read(rnd[:])
	tmp := filepath.Join(filepath.Dir(dir), "."+filepath.Base(dir)+".partial-"+hex.EncodeToString(rnd[:]))
	if err := os.Mkdir(tmp, 0o755); err != nil {
		return "", err
	}
	ok := false
	defer func() {
		if !ok {
			os.RemoveAll(tmp)
		}
	}()
	sortRecords(&recs)
	ident.Datasets = nil
	for _, df := range datasetFiles {
		ds, err := writeDataset(filepath.Join(tmp, df.path), recs.of(df.recordType))
		if err != nil {
			return "", err
		}
		ds.Path = df.path
		ident.Datasets = append(ident.Datasets, ds)
	}
	idBytes, err := json.MarshalIndent(ident, "", "  ")
	if err != nil {
		return "", err
	}
	idBytes = append(idBytes, '\n')
	if err := writeSynced(filepath.Join(tmp, "identity.json"), idBytes); err != nil {
		return "", err
	}
	sum := sha256.Sum256(idBytes)
	analysisID := "sha256:" + hex.EncodeToString(sum[:])
	man := Manifest{SchemaVersion: BundleSchema, ContractRevision: ContractRevision, AnalysisID: analysisID,
		IdentityPath: "identity.json", CreatedAt: createdAt.UTC().Format(time.RFC3339)}
	mb, _ := json.MarshalIndent(man, "", "  ")
	if err := writeSynced(filepath.Join(tmp, "manifest.json"), append(mb, '\n')); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dir); err != nil {
		return "", err
	}
	ok = true
	return analysisID, nil
}

func writeSynced(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func writeDataset(path string, recs []any) (Dataset, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return Dataset{}, err
	}
	defer f.Close()
	h := sha256.New()
	w := bufio.NewWriterSize(io.MultiWriter(f, h), 1<<20)
	var size int64
	for _, r := range recs {
		b, err := json.Marshal(r)
		if err != nil {
			return Dataset{}, err
		}
		b = append(b, '\n')
		if _, err := w.Write(b); err != nil {
			return Dataset{}, err
		}
		size += int64(len(b))
	}
	if err := w.Flush(); err != nil {
		return Dataset{}, err
	}
	if err := f.Sync(); err != nil {
		return Dataset{}, err
	}
	return Dataset{SHA256: "sha256:" + hex.EncodeToString(h.Sum(nil)), SizeBytes: size, RecordCount: len(recs), Encoding: "jsonl-utf8"}, nil
}

func (r *Records) of(t string) []any {
	var out []any
	switch t {
	case TypeComponent:
		for _, x := range r.Components {
			out = append(out, x)
		}
	case TypeClass:
		for _, x := range r.Classes {
			out = append(out, x)
		}
	case TypeNode:
		for _, x := range r.Nodes {
			out = append(out, x)
		}
	case TypeEdge:
		for _, x := range r.Edges {
			out = append(out, x)
		}
	case TypeRoot:
		for _, x := range r.Roots {
			out = append(out, x)
		}
	case TypeGap:
		for _, x := range r.Gaps {
			out = append(out, x)
		}
	case TypeCall:
		for _, x := range r.Calls {
			out = append(out, x)
		}
	}
	return out
}

func sortRecords(r *Records) {
	sort.Slice(r.Components, func(i, j int) bool { return r.Components[i].InstanceID < r.Components[j].InstanceID })
	sort.Slice(r.Classes, func(i, j int) bool { return r.Classes[i].ClassID < r.Classes[j].ClassID })
	sort.Slice(r.Nodes, func(i, j int) bool { return r.Nodes[i].NodeID < r.Nodes[j].NodeID })
	sort.Slice(r.Edges, func(i, j int) bool { return r.Edges[i].EdgeID < r.Edges[j].EdgeID })
	sort.Slice(r.Roots, func(i, j int) bool { return r.Roots[i].RootID < r.Roots[j].RootID })
	sort.Slice(r.Gaps, func(i, j int) bool { return r.Gaps[i].GapID < r.Gaps[j].GapID })
	sort.Slice(r.Calls, func(i, j int) bool { return r.Calls[i].CallID < r.Calls[j].CallID })
}

var datasetPathRE = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)

// safeMember resolves a bundle-relative path, rejecting absolute paths,
// "..", backslashes and a symlink anywhere below dir (CONTRACT.md §4).
func safeMember(dir, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") || strings.Contains(rel, `\`) || !datasetPathRE.MatchString(rel) {
		return "", corrupt("bundle path %q is not a relative path below the bundle", rel)
	}
	cur := dir
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." || part == ".." {
			return "", corrupt("bundle path %q is not a relative path below the bundle", rel)
		}
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if err != nil {
			return "", corrupt("bundle member %s: %v", rel, err)
		}
		if fi.Mode()&fs.ModeSymlink != 0 {
			return "", corrupt("bundle member %s is reached through a symlink", rel)
		}
	}
	return cur, nil
}

func readLimited(path string, limit int64) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Size() > limit {
		return nil, fmt.Errorf("%s is %d bytes, over the reader's limit of %d", path, fi.Size(), limit)
	}
	return os.ReadFile(path)
}

// Bundle is a loaded, verified bundle.
type Bundle struct {
	Dir        string
	Manifest   Manifest
	Identity   Identity
	AnalysisID string
	Records
	idx     index
	loaded  map[string]bool // record types this bundle was loaded with
	skipped []string        // datasets checked for presence and size only
	mu      sync.Mutex
}

// Load reads and verifies a whole bundle; see LoadFor.
func Load(dir string) (*Bundle, error) { return LoadFor(dir, "") }

// DatasetsFor names the record types an operation reads ("" or an
// unknown operation: every type).
func DatasetsFor(op string) map[string]bool {
	set := func(ts ...string) map[string]bool {
		m := map[string]bool{}
		for _, t := range ts {
			m[t] = true
		}
		return m
	}
	switch op {
	case "entries":
		return set(TypeComponent, TypeNode, TypeRoot)
	case "methods", "callers", "callees":
		return set(TypeComponent, TypeNode, TypeEdge, TypeRoot)
	case "calls":
		return set(TypeComponent, TypeNode, TypeEdge, TypeRoot, TypeCall)
	case "reachability":
		return set(TypeComponent, TypeClass, TypeNode, TypeEdge, TypeRoot, TypeGap)
	}
	return set(TypeComponent, TypeClass, TypeNode, TypeEdge, TypeRoot, TypeGap, TypeCall)
}

// LoadFor reads and verifies a bundle for one operation (CONTRACT.md §4,
// §5): format and revision, the identity hash, every dataset's presence
// and size; and, for the datasets the operation reads, the hash, record
// count, strict record shapes and every cross-reference among them.
func LoadFor(dir, op string) (*Bundle, error) {
	fi, err := os.Lstat(dir)
	if err != nil {
		return nil, &Error{Code: "IO_ERROR", Message: err.Error()}
	}
	if fi.Mode()&fs.ModeSymlink != 0 || !fi.IsDir() {
		return nil, corrupt("%s is not a bundle directory", dir)
	}
	mpath, err := safeMember(dir, "manifest.json")
	if err != nil {
		return nil, err
	}
	mb, err := readLimited(mpath, maxManifestBytes)
	if err != nil {
		return nil, corrupt("manifest: %v", err)
	}
	b := &Bundle{Dir: dir, loaded: DatasetsFor(op)}
	if err := decodeStrict(mb, &b.Manifest); err != nil {
		return nil, corrupt("manifest.json: %v", err)
	}
	if b.Manifest.SchemaVersion != BundleSchema || b.Manifest.ContractRevision != ContractRevision {
		return nil, &Error{Code: "INCOMPATIBLE_BUNDLE", Message: fmt.Sprintf("bundle %s revision %s; this reader supports %s %s",
			b.Manifest.SchemaVersion, b.Manifest.ContractRevision, BundleSchema, ContractRevision)}
	}
	ipath, err := safeMember(dir, b.Manifest.IdentityPath)
	if err != nil {
		return nil, err
	}
	ib, err := readLimited(ipath, maxIdentityBytes)
	if err != nil {
		return nil, corrupt("identity: %v", err)
	}
	sum := sha256.Sum256(ib)
	b.AnalysisID = "sha256:" + hex.EncodeToString(sum[:])
	if b.AnalysisID != b.Manifest.AnalysisID {
		return nil, corrupt("identity hash %s does not match the manifest's analysis_id %s", b.AnalysisID, b.Manifest.AnalysisID)
	}
	if err := decodeStrict(ib, &b.Identity); err != nil {
		return nil, corrupt("%s: %v", b.Manifest.IdentityPath, err)
	}
	id := b.Identity
	if id.SchemaVersion != IdentitySchema || id.ContractRevision != ContractRevision {
		return nil, &Error{Code: "INCOMPATIBLE_BUNDLE", Message: fmt.Sprintf("identity %s revision %s is not supported", id.SchemaVersion, id.ContractRevision)}
	}
	if err := checkIdentity(id); err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	errs := make([]error, len(id.Datasets))
	for i, ds := range id.Datasets {
		wg.Add(1)
		go func(i int, ds Dataset) {
			defer wg.Done()
			errs[i] = b.readDataset(ds)
		}(i, ds)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	if err := b.buildIndex(); err != nil {
		return nil, err
	}
	return b, nil
}

func checkIdentity(id Identity) error {
	if len(id.Inputs) == 0 || len(id.Datasets) == 0 {
		return corrupt("identity lists no inputs or no datasets")
	}
	if len(id.Datasets) > maxDatasets {
		return corrupt("identity lists %d datasets, over the reader's limit of %d", len(id.Datasets), maxDatasets)
	}
	inputs := map[string]bool{}
	for _, in := range id.Inputs {
		if !identifierRE.MatchString(in.ID) || inputs[in.ID] {
			return corrupt("input id %q is invalid or repeated", in.ID)
		}
		inputs[in.ID] = true
	}
	for _, c := range id.ClasspathOrder {
		if !inputs[c] {
			return corrupt("classpath_order names unknown input %q", c)
		}
	}
	paths := map[string]bool{}
	for _, ds := range id.Datasets {
		if paths[ds.Path] {
			return corrupt("dataset %s listed twice", ds.Path)
		}
		paths[ds.Path] = true
		if ds.Encoding != "jsonl-utf8" {
			return &Error{Code: "INCOMPATIBLE_BUNDLE", Message: fmt.Sprintf("dataset %s encoding %q is not supported", ds.Path, ds.Encoding)}
		}
		if ds.SizeBytes < 0 || ds.SizeBytes > maxDatasetBytes {
			return corrupt("dataset %s declares %d bytes, outside the reader's limit", ds.Path, ds.SizeBytes)
		}
	}
	for _, c := range id.Capabilities {
		if !contains(Capabilities, c) {
			return &Error{Code: "INCOMPATIBLE_BUNDLE", Message: fmt.Sprintf("bundle capability %q is not supported", c)}
		}
	}
	return nil
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// required lists the members every record of a type must carry, nullable
// ones included (the schema's "required").
var required = map[string][]string{
	TypeComponent: {"record_type", "instance_id", "role", "identity", "gav", "artifact_sha256", "location"},
	TypeClass:     {"record_type", "class_id", "instance_id", "class", "class_sha256", "selected", "selected_class_id", "variant_java"},
	TypeNode:      {"record_type", "node_id", "node_kind", "method", "canonical_node_id", "receiver_type", "body", "access_flags", "source"},
	TypeEdge:      {"record_type", "edge_id", "from", "to", "kind", "basis", "character", "rule", "sites", "client_dispatch", "supplied_dispatch"},
	TypeRoot:      {"record_type", "root_id", "node_id", "basis", "inherited_by", "announced_by"},
	TypeGap:       {"record_type", "gap_id", "mechanism", "reason", "site", "reachable_from_indexed_roots", "candidate_method_ids", "candidate_instance_ids", "rule"},
	TypeCall:      {"record_type", "call_id", "site", "declared_class", "declared_name", "declared_descriptor", "target_node_ids", "body_in_input"},
}

// readDataset checks a dataset's presence and size and, unless the
// operation does not read it, streams it: hash, record count and every
// record decoded and validated in parallel, without holding the file in
// memory. A dataset is skipped (presence and size only) when it is one of
// the per-type files this producer writes (datasetFiles), its first record
// has that type, and the operation does not read the type.
func (b *Bundle) readDataset(ds Dataset) error {
	path, err := safeMember(b.Dir, ds.Path)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return corrupt("dataset %s: %v", ds.Path, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return &Error{Code: "IO_ERROR", Message: err.Error()}
	}
	if fi.Size() != ds.SizeBytes {
		return corrupt("dataset %s is %d bytes, the identity declares %d", ds.Path, fi.Size(), ds.SizeBytes)
	}
	h := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(f, h), 1<<20)
	if first, err := r.Peek(min(4096, int(ds.SizeBytes))); err == nil || len(first) > 0 {
		if nl := bytes.IndexByte(first, '\n'); nl > 0 {
			t := recordTypeOf(first[:nl])
			for _, df := range datasetFiles {
				if df.path == ds.Path && df.recordType == t && !b.loaded[t] {
					b.mu.Lock()
					b.skipped = append(b.skipped, ds.Path)
					b.mu.Unlock()
					return nil
				}
			}
		}
	}

	// Batches of lines go to workers; results are reassembled in order.
	const batchSize = 4096
	type batch struct {
		idx   int
		first int // line number of the first line, 1-based
		lines [][]byte
	}
	type result struct {
		recs []any
		err  error
	}
	workers := runtime.GOMAXPROCS(0)
	in := make(chan batch, workers*2)
	var mu sync.Mutex
	results := map[int]result{}
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for bt := range in {
				recs := make([]any, len(bt.lines))
				var err error
				for i, l := range bt.lines {
					if recs[i], err = decodeRecord(l); err != nil {
						err = fmt.Errorf("line %d: %v", bt.first+i, err)
						break
					}
				}
				mu.Lock()
				results[bt.idx] = result{recs, err}
				mu.Unlock()
			}
		}()
	}
	n, nb := 0, 0
	cur := batch{first: 1}
	var readErr error
	for {
		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			if len(line) > 0 {
				readErr = corrupt("dataset %s does not end with LF", ds.Path)
			}
			break
		}
		if err != nil {
			readErr = &Error{Code: "IO_ERROR", Message: err.Error()}
			break
		}
		n++
		line = line[:len(line)-1]
		if len(line) == 0 {
			readErr = corrupt("dataset %s line %d is empty", ds.Path, n)
			break
		}
		if len(line) > maxLineBytes {
			readErr = corrupt("dataset %s line %d is over the reader's limit", ds.Path, n)
			break
		}
		cur.lines = append(cur.lines, line)
		if len(cur.lines) == batchSize {
			cur.idx = nb
			in <- cur
			nb++
			cur = batch{first: n + 1}
		}
	}
	if readErr == nil && len(cur.lines) > 0 {
		cur.idx = nb
		in <- cur
		nb++
	}
	close(in)
	wg.Wait()
	if readErr != nil {
		return readErr
	}
	if got := "sha256:" + hex.EncodeToString(h.Sum(nil)); got != ds.SHA256 {
		return corrupt("dataset %s hash %s does not match the identity's %s", ds.Path, got, ds.SHA256)
	}
	if n != ds.RecordCount {
		return corrupt("dataset %s has %d records, the identity declares %d", ds.Path, n, ds.RecordCount)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := 0; i < nb; i++ {
		res := results[i]
		if res.err != nil {
			return corrupt("dataset %s %v", ds.Path, res.err)
		}
		for _, rec := range res.recs {
			switch r := rec.(type) {
			case Component:
				b.Components = append(b.Components, r)
			case Class:
				b.Classes = append(b.Classes, r)
			case Node:
				b.Nodes = append(b.Nodes, r)
			case Edge:
				b.Edges = append(b.Edges, r)
			case Root:
				b.Roots = append(b.Roots, r)
			case Gap:
				b.Gaps = append(b.Gaps, r)
			case Call:
				b.Calls = append(b.Calls, r)
			}
		}
	}
	return nil
}

// recordTypeOf is a cheap look at a line's record_type ("" if none).
func recordTypeOf(line []byte) string {
	i := bytes.Index(line, []byte(`"record_type":"`))
	if i < 0 {
		return ""
	}
	rest := line[i+len(`"record_type":"`):]
	j := bytes.IndexByte(rest, '"')
	if j < 0 {
		return ""
	}
	return string(rest[:j])
}

// decodeRecord validates one record's members on the raw bytes — valid
// UTF-8, and for the record and each nested object exactly the members the
// schema defines (every one of them is required, nullable ones included):
// no duplicate, unknown or missing member — then decodes it.
func decodeRecord(line []byte) (any, error) {
	typ := recordTypeOf(line)
	if _, ok := required[typ]; !ok {
		return nil, fmt.Errorf("unknown record_type %q", typ)
	}
	if !utf8.Valid(line) {
		return nil, fmt.Errorf("record is not valid UTF-8")
	}
	if err := checkMembers(line, typ); err != nil {
		return nil, err
	}
	var err error
	var v any
	switch typ {
	case TypeComponent:
		var r Component
		err, v = json.Unmarshal(line, &r), &r
	case TypeClass:
		var r Class
		err, v = json.Unmarshal(line, &r), &r
	case TypeNode:
		var r Node
		err, v = json.Unmarshal(line, &r), &r
	case TypeEdge:
		var r Edge
		err, v = json.Unmarshal(line, &r), &r
	case TypeRoot:
		var r Root
		err, v = json.Unmarshal(line, &r), &r
	case TypeGap:
		var r Gap
		err, v = json.Unmarshal(line, &r), &r
	case TypeCall:
		var r Call
		err, v = json.Unmarshal(line, &r), &r
	}
	if err != nil {
		return nil, err
	}
	switch r := v.(type) {
	case *Component:
		return *r, nil
	case *Class:
		return *r, nil
	case *Node:
		return *r, nil
	case *Edge:
		return *r, nil
	case *Root:
		return *r, nil
	case *Gap:
		return *r, nil
	case *Call:
		return *r, nil
	}
	return nil, fmt.Errorf("record_type %q does not match its record", typ)
}

// nestedRequired lists the members required in a record's nested objects,
// by the top-level member holding them (an array holds objects of that
// shape).
var nestedRequired = map[string]map[string][]string{
	TypeNode: {"method": {"method_id", "instance_id", "class_sha256", "class", "name", "descriptor"},
		"source": {"path", "line_start", "line_end", "source_sha256", "binding"}},
	TypeEdge: {"sites": siteMembers},
	TypeGap:  {"site": siteMembers},
	TypeCall: {"site": siteMembers},
}

var siteMembers = []string{"caller_node_id", "pc", "kind", "condition"}

// member is one object member: its key as written and its raw value.
type member struct{ key, val []byte }

// checkMembers verifies the member sets by scanning the raw bytes.
func checkMembers(line []byte, typ string) error {
	var buf [16]member
	top, err := objectMembers(line, buf[:0])
	if err != nil {
		return err
	}
	if err := exactMembers(top, required[typ], ""); err != nil {
		return err
	}
	if v := find(top, "record_type"); len(v) != len(typ)+2 || string(v[1:len(v)-1]) != typ {
		return fmt.Errorf("record_type member does not match %q", typ)
	}
	var nbuf [8]member
	for name, want := range nestedRequired[typ] {
		raw := find(top, name)
		if len(raw) == 0 || raw[0] == 'n' { // null: allowed where nullable; the decode checks the type
			continue
		}
		objs := [][]byte{raw}
		if raw[0] == '[' {
			if objs, err = arrayElements(raw); err != nil {
				return err
			}
		}
		for _, o := range objs {
			m, err := objectMembers(o, nbuf[:0])
			if err != nil {
				return err
			}
			if err := exactMembers(m, want, "/"+name); err != nil {
				return err
			}
		}
	}
	return nil
}

func find(ms []member, name string) []byte {
	for _, m := range ms {
		if string(m.key) == name {
			return m.val
		}
	}
	return nil
}

// exactMembers: ms has exactly the members names, each once (duplicates
// were rejected by objectMembers).
func exactMembers(ms []member, names []string, at string) error {
	for _, n := range names {
		if find(ms, n) == nil {
			return fmt.Errorf("missing required field %q at %s", n, pointerOrRoot(at))
		}
	}
	if len(ms) != len(names) {
		for _, m := range ms {
			if !contains(names, string(m.key)) {
				return fmt.Errorf("unknown field %q at %s", m.key, pointerOrRoot(at))
			}
		}
	}
	return nil
}

// objectMembers splits a JSON object into its members, appended to out,
// rejecting a repeated member; it checks structure only as far as needed
// to find them (the decode validates the rest). Keys are compared as
// written: a key with an escape sequence never equals a schema name.
func objectMembers(b []byte, out []member) ([]member, error) {
	b = bytes.TrimSpace(b)
	if len(b) < 2 || b[0] != '{' {
		return nil, fmt.Errorf("not a JSON object")
	}
	i := skipWS(b, 1)
	if i < len(b) && b[i] == '}' {
		return out, nil
	}
	for i < len(b) {
		if b[i] != '"' {
			return nil, fmt.Errorf("malformed object")
		}
		end, err := skipString(b, i)
		if err != nil {
			return nil, err
		}
		key := b[i+1 : end-1]
		for _, m := range out {
			if bytes.Equal(m.key, key) {
				return nil, fmt.Errorf("duplicate member %q", key)
			}
		}
		i = skipWS(b, end)
		if i >= len(b) || b[i] != ':' {
			return nil, fmt.Errorf("malformed object")
		}
		i = skipWS(b, i+1)
		vend, err := skipValue(b, i)
		if err != nil {
			return nil, err
		}
		out = append(out, member{key, b[i:vend]})
		i = skipWS(b, vend)
		if i < len(b) && b[i] == ',' {
			i = skipWS(b, i+1)
			continue
		}
		if i < len(b) && b[i] == '}' {
			return out, nil
		}
		return nil, fmt.Errorf("malformed object")
	}
	return nil, fmt.Errorf("unterminated object")
}

func arrayElements(b []byte) ([][]byte, error) {
	var out [][]byte
	i := skipWS(b, 1)
	if i < len(b) && b[i] == ']' {
		return out, nil
	}
	for i < len(b) {
		end, err := skipValue(b, i)
		if err != nil {
			return nil, err
		}
		out = append(out, b[i:end])
		i = skipWS(b, end)
		if i < len(b) && b[i] == ',' {
			i = skipWS(b, i+1)
			continue
		}
		if i < len(b) && b[i] == ']' {
			return out, nil
		}
		return nil, fmt.Errorf("malformed array")
	}
	return nil, fmt.Errorf("unterminated array")
}

func skipWS(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n' || b[i] == '\r') {
		i++
	}
	return i
}

// skipString returns the index after the string starting at b[i] == '"'.
func skipString(b []byte, i int) (int, error) {
	for j := i + 1; j < len(b); j++ {
		switch b[j] {
		case '\\':
			j++
		case '"':
			return j + 1, nil
		}
	}
	return 0, fmt.Errorf("unterminated string")
}

// skipValue returns the index after the JSON value starting at b[i].
func skipValue(b []byte, i int) (int, error) {
	if i >= len(b) {
		return 0, fmt.Errorf("missing value")
	}
	switch b[i] {
	case '"':
		return skipString(b, i)
	case '{', '[':
		depth := 0
		for j := i; j < len(b); j++ {
			switch b[j] {
			case '"':
				end, err := skipString(b, j)
				if err != nil {
					return 0, err
				}
				j = end - 1
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					return j + 1, nil
				}
			}
		}
		return 0, fmt.Errorf("unterminated value")
	}
	j := i
	for j < len(b) && b[j] != ',' && b[j] != '}' && b[j] != ']' && b[j] != ' ' && b[j] != '\n' {
		j++
	}
	if j == i {
		return 0, fmt.Errorf("missing value")
	}
	return j, nil
}
