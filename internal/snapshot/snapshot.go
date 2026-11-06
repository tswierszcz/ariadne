// Package snapshot records and replays the data an analysis used
// (CVE:F3): a manifest plus a content-addressed object store.
//
// A snapshot reproduces the *data* of a run — every cache lookup outcome
// (OSV, fix commits, LLM answers, misses and errors included), the input
// artifacts' hashes, the configuration and the analyzer's tool and
// algorithm versions. It does not reproduce the analyzer itself: a replay
// requires the same tool version and algorithm versions, and for a
// locally modified ("dirty") build the same binary. Input artifacts are
// not archived; the manifest names them and a replay checks their hashes.
//
// Layout:
//
//	<dir>/manifest.json
//	<dir>/objects/<sha256[:2]>/<sha256>
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ariadne/internal/atomicfile"
	"ariadne/internal/httpcache"
)

// FormatVersion is the manifest format; a replay refuses another one.
const FormatVersion = 1

// Manifest describes one snapshot.
type Manifest struct {
	FormatVersion int `json:"format_version"`
	// DataSetID identifies the logical data set: a SHA-256 over the
	// sorted observations (not over the creation time), so two runs that
	// read the same data have the same ID.
	DataSetID  string            `json:"data_set_id"`
	CreatedAt  time.Time         `json:"created_at"`
	Tool       Tool              `json:"tool"`
	Algorithms map[string]string `json:"algorithms"`
	// Config is the analysis configuration (flag values that shape the
	// result; output paths and secrets excluded).
	Config       map[string]string       `json:"config"`
	Inputs       []Input                 `json:"inputs"`
	Observations []httpcache.Observation `json:"observations"`
	Derivations  []Derivation            `json:"derivations"`
	Verdicts     map[string]string       `json:"verdicts"` // advisory -> v1 status
	// DataFreshness is the freshness state of the recorded run (CVE:F2).
	DataFreshness string `json:"data_freshness,omitempty"`
	// AnalysisID is the analysis_id of the results the recording run
	// wrote with this manifest (CVE:F9.9).
	AnalysisID string `json:"analysis_id,omitempty"`
}

// Tool identifies the analyzer build that made the snapshot.
type Tool struct {
	Version      string `json:"version"`
	GoVersion    string `json:"go_version"`
	BinarySHA256 string `json:"binary_sha256,omitempty"`
	Dirty        bool   `json:"dirty"`
	// Reproducible is false for a dirty build: the source of that binary
	// is not identified by its version, only the binary hash is.
	Reproducible bool   `json:"reproducible"`
	Note         string `json:"note,omitempty"`
}

// Input is one analyzed input file, identified by stable location.
type Input struct {
	Role     string `json:"role"` // artifact role, "deps-list", "vuln-methods"
	Location string `json:"location"`
	SHA256   string `json:"sha256"`
}

// Derivation is the key of one derived mapping (CVE_IMPLEMENTATION_PLAN
// F3.4): the advisory content, the affected component instances and
// versions, the mapping's sources and the resolver/prompt versions. A
// change of any of them changes Key and invalidates the mapping.
type Derivation struct {
	Advisory     string         `json:"advisory"`
	AdvisoryHash string         `json:"advisory_hash"`
	Instances    []string       `json:"instances,omitempty"` // affected instances by stable artifact location
	Affected     []string       `json:"affected"`
	Source       string         `json:"source,omitempty"` // manual, diff, llm; "" = no mapping
	MappingHash  string         `json:"mapping_hash"`     // methods + role/coverage/review
	SourceHashes []string       `json:"source_hashes,omitempty"`
	Resolver     string         `json:"resolver"`
	LLM          *LLMDerivation `json:"llm,omitempty"`
	Key          string         `json:"key"`
}

// LLMDerivation adds the LLM-specific parts of a derivation key.
type LLMDerivation struct {
	Model         string `json:"model"`
	PromptVersion string `json:"prompt_version"`
	// Reused is set when the answer was given for an earlier version of
	// the advisory (CVE:F2 reuse rule): the derivation is stale.
	Reused bool `json:"reused,omitempty"`
}

// HashJSON is the SHA-256 of v's JSON encoding.
func HashJSON(v any) string {
	b, _ := json.Marshal(v)
	return HashBytes(b)
}

// HashBytes is a hex SHA-256.
func HashBytes(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// ComputeKey fills d.Key from every other field.
func (d *Derivation) ComputeKey() {
	c := *d
	c.Key = ""
	d.Key = HashJSON(c)
}

// Recorder collects observations and writes hit bodies into the object
// store as they happen. Safe for concurrent use.
type Recorder struct {
	dir string
	mu  sync.Mutex
	obs map[string]httpcache.Observation // namespace/key -> final observation
	err error
}

// NewRecorder creates dir (which must not already hold a manifest).
func NewRecorder(dir string) (*Recorder, error) {
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err == nil {
		return nil, fmt.Errorf("snapshot %s already exists (manifest.json); choose a new directory", dir)
	}
	if err := os.MkdirAll(filepath.Join(dir, "objects"), 0o755); err != nil {
		return nil, fmt.Errorf("creating snapshot dir: %w", err)
	}
	return &Recorder{dir: dir, obs: map[string]httpcache.Observation{}}, nil
}

// Observe implements httpcache.Recorder. A hit overrides an earlier miss
// or error of the same key (the final state is what a replay needs: an
// LLM answer is a miss, then a Set).
func (r *Recorder) Observe(o httpcache.Observation, body []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := o.Namespace + "/" + o.Key
	if prev, ok := r.obs[id]; ok && prev.Outcome == httpcache.OutcomeHit && o.Outcome != httpcache.OutcomeHit {
		return
	}
	if o.Outcome == httpcache.OutcomeHit {
		if err := writeObject(r.dir, o.SHA256, body); err != nil && r.err == nil {
			r.err = err
		}
	}
	r.obs[id] = o
}

func objectPath(dir, sum string) string {
	return filepath.Join(dir, "objects", sum[:2], sum)
}

func writeObject(dir, sum string, body []byte) error {
	p := objectPath(dir, sum)
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), sum+".*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(body)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("writing snapshot object %s: %v %v", sum, werr, cerr)
	}
	return os.Rename(tmp.Name(), p)
}

// Observations returns the recorded observations sorted by namespace and
// key, and the first object-write error.
func (r *Recorder) Observations() ([]httpcache.Observation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]httpcache.Observation, 0, len(r.obs))
	for _, o := range r.obs {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Key < out[j].Key
	})
	return out, r.err
}

// DataSetID hashes the observations' identity and content (key, outcome,
// body hash, error) — not their fetch times.
func DataSetID(obs []httpcache.Observation) string {
	lines := make([]string, len(obs))
	for i, o := range obs {
		lines[i] = strings.Join([]string{o.Namespace, o.Key, o.Outcome, o.SHA256, o.ErrClass, o.Error}, "\t")
	}
	return "sha256:" + HashBytes([]byte(strings.Join(lines, "\n")))
}

// ManifestPath is dir's manifest file.
func ManifestPath(dir string) string { return filepath.Join(dir, "manifest.json") }

// Marshal renders m as manifest JSON, setting its format and DataSetID.
func Marshal(m Manifest) ([]byte, error) {
	m.FormatVersion = FormatVersion
	m.DataSetID = DataSetID(m.Observations)
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling snapshot manifest: %w", err)
	}
	return append(b, '\n'), nil
}

// Write writes m as dir/manifest.json, setting DataSetID, atomically.
func Write(dir string, m Manifest) error {
	b, err := Marshal(m)
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(ManifestPath(dir), b)
}

// Read loads dir/manifest.json.
func Read(dir string) (Manifest, error) {
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("reading snapshot manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, fmt.Errorf("decoding snapshot manifest: %w", err)
	}
	if m.FormatVersion != FormatVersion {
		return Manifest{}, fmt.Errorf("snapshot format %d, this analyzer reads format %d", m.FormatVersion, FormatVersion)
	}
	if got := DataSetID(m.Observations); got != m.DataSetID {
		return Manifest{}, fmt.Errorf("snapshot manifest data set ID %s does not match its observations (%s): edited manifest", m.DataSetID, got)
	}
	return m, nil
}

// ErrCorrupt is returned when an object does not match its recorded hash.
var ErrCorrupt = errors.New("snapshot object does not match its recorded SHA-256")

// Replayer serves a snapshot's observations (httpcache.Replayer).
type Replayer struct {
	dir string
	obs map[string]httpcache.Observation

	mu       sync.Mutex
	failures []string
}

// Failures lists every lookup the snapshot could not answer (not
// recorded, or an object that fails its hash), sorted. A caller may
// swallow such an error as a per-CVE warning, so the run must check this
// before trusting a replayed result.
func (r *Replayer) Failures() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := append([]string(nil), r.failures...)
	sort.Strings(out)
	return out
}

func (r *Replayer) fail(msg string) {
	r.mu.Lock()
	r.failures = append(r.failures, msg)
	r.mu.Unlock()
}

// NewReplayer indexes m's observations; bodies are read from dir lazily
// and verified on every read.
func NewReplayer(dir string, m Manifest) *Replayer {
	r := &Replayer{dir: dir, obs: map[string]httpcache.Observation{}}
	for _, o := range m.Observations {
		r.obs[o.Namespace+"/"+o.Key] = o
	}
	return r
}

// Lookup implements httpcache.Replayer.
func (r *Replayer) Lookup(namespace, key string) (httpcache.Observation, []byte, bool, error) {
	o, ok := r.obs[namespace+"/"+key]
	if !ok {
		r.fail(fmt.Sprintf("%s/%s: not in the snapshot", namespace, key))
		return o, nil, false, nil
	}
	if o.Outcome != httpcache.OutcomeHit {
		return o, nil, true, nil
	}
	b, err := os.ReadFile(objectPath(r.dir, o.SHA256))
	if err != nil {
		err = fmt.Errorf("snapshot object %s for %s/%s: %w", o.SHA256, namespace, key, err)
		r.fail(err.Error())
		return o, nil, true, err
	}
	if HashBytes(b) != o.SHA256 {
		err := fmt.Errorf("%w: %s/%s (object %s)", ErrCorrupt, namespace, key, o.SHA256)
		r.fail(err.Error())
		return o, nil, true, err
	}
	return o, b, true, nil
}

// CheckTool refuses a replay by another analyzer: the tool version and
// every algorithm version must match, and for a dirty build the binary.
func CheckTool(m Manifest, tool Tool, algorithms map[string]string) error {
	if m.Tool.Version != tool.Version {
		return fmt.Errorf("snapshot made by ariadne %s, this is %s: a replay needs the same analyzer version", m.Tool.Version, tool.Version)
	}
	if m.Tool.Dirty && m.Tool.BinarySHA256 != tool.BinarySHA256 {
		return fmt.Errorf("snapshot made by a locally modified build (%s) whose binary differs from this one: the data can be replayed only with that binary", m.Tool.Version)
	}
	var diffs []string
	for k, v := range m.Algorithms {
		if algorithms[k] != v {
			diffs = append(diffs, fmt.Sprintf("%s %s → %s", k, v, algorithms[k]))
		}
	}
	for k, v := range algorithms {
		if _, ok := m.Algorithms[k]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s (new) %s", k, v))
		}
	}
	if len(diffs) > 0 {
		sort.Strings(diffs)
		return fmt.Errorf("snapshot algorithm versions differ: %s", strings.Join(diffs, "; "))
	}
	return nil
}

// CheckInputs compares the run's inputs with the manifest's by location:
// a changed, missing or extra input is an error.
func CheckInputs(m Manifest, got []Input) error {
	want := map[string]Input{}
	for _, in := range m.Inputs {
		want[in.Role+"\t"+in.Location] = in
	}
	var problems []string
	seen := map[string]bool{}
	for _, in := range got {
		id := in.Role + "\t" + in.Location
		seen[id] = true
		w, ok := want[id]
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("%s %s is not in the snapshot", in.Role, in.Location))
		case w.SHA256 != in.SHA256:
			problems = append(problems, fmt.Sprintf("%s %s has SHA-256 %s, the snapshot recorded %s", in.Role, in.Location, short(in.SHA256), short(w.SHA256)))
		}
	}
	for id, in := range want {
		if !seen[id] {
			problems = append(problems, fmt.Sprintf("%s %s recorded in the snapshot is missing", in.Role, in.Location))
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return fmt.Errorf("inputs do not match the snapshot manifest: %s", strings.Join(problems, "; "))
	}
	return nil
}

func short(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

// ConfigDiff lists configuration keys whose value differs.
func ConfigDiff(old, cur map[string]string) []string {
	keys := map[string]bool{}
	for k := range old {
		keys[k] = true
	}
	for k := range cur {
		keys[k] = true
	}
	var out []string
	for k := range keys {
		if old[k] != cur[k] {
			out = append(out, fmt.Sprintf("%s: %q → %q", k, old[k], cur[k]))
		}
	}
	sort.Strings(out)
	return out
}

// Change is one advisory whose derived result is invalidated, or whose
// verdict changed, between an earlier snapshot and this run.
type Change struct {
	Advisory string   `json:"advisory"`
	Reasons  []string `json:"reasons"`
	Verdict  string   `json:"verdict,omitempty"` // "OLD → NEW" when it changed
}

// Compare explains, per advisory, what changed since old: new or gone
// advisories, advisory content, affected instances, mapping, mapping
// sources, resolver or LLM versions, input artifacts, algorithm versions
// and the v1 verdict.
func Compare(old Manifest, cur Manifest) []Change {
	oldD := map[string]Derivation{}
	for _, d := range old.Derivations {
		oldD[d.Advisory] = d
	}
	var global []string
	for _, a := range ConfigDiff(old.Algorithms, cur.Algorithms) {
		global = append(global, "algorithm "+a)
	}
	if old.Tool.Version != cur.Tool.Version {
		global = append(global, fmt.Sprintf("analyzer %s → %s", old.Tool.Version, cur.Tool.Version))
	}
	for _, c := range ConfigDiff(old.Config, cur.Config) {
		global = append(global, "config "+c)
	}
	inputChanged := map[string]bool{}
	oldIn := map[string]string{}
	for _, in := range old.Inputs {
		oldIn[in.Role+"\t"+in.Location] = in.SHA256
	}
	for _, in := range cur.Inputs {
		if h, ok := oldIn[in.Role+"\t"+in.Location]; !ok || h != in.SHA256 {
			inputChanged[in.Location] = true
		}
	}

	var out []Change
	seen := map[string]bool{}
	for _, d := range cur.Derivations {
		seen[d.Advisory] = true
		o, ok := oldD[d.Advisory]
		var reasons []string
		if !ok {
			reasons = append(reasons, "new advisory")
		} else if o.Key != d.Key || len(global) > 0 {
			if o.AdvisoryHash != d.AdvisoryHash {
				reasons = append(reasons, "advisory content changed")
			}
			if strings.Join(o.Affected, ",") != strings.Join(d.Affected, ",") || strings.Join(o.Instances, ",") != strings.Join(d.Instances, ",") {
				reasons = append(reasons, "affected component instances changed")
			}
			if o.Source != d.Source || o.MappingHash != d.MappingHash {
				reasons = append(reasons, fmt.Sprintf("mapping changed (%s → %s)", orNone(o.Source), orNone(d.Source)))
			}
			if strings.Join(o.SourceHashes, ",") != strings.Join(d.SourceHashes, ",") {
				reasons = append(reasons, "mapping sources (fix commits, LLM answer) changed")
			}
			if o.Resolver != d.Resolver {
				reasons = append(reasons, "resolver version changed")
			}
			if HashJSON(o.LLM) != HashJSON(d.LLM) {
				reasons = append(reasons, "LLM model, prompt version or answer freshness changed")
			}
			reasons = append(reasons, global...)
		}
		if d.LLM != nil && d.LLM.Reused {
			reasons = append(reasons, "LLM answer reused from an earlier advisory version (stale derivation)")
		}
		for _, in := range d.Instances {
			if inputChanged[in] {
				reasons = append(reasons, "input artifact changed")
				break
			}
		}
		verdict := ""
		if ov, nv := old.Verdicts[d.Advisory], cur.Verdicts[d.Advisory]; !ok {
			verdict = "(none) → " + nv
		} else if ov != nv {
			verdict = ov + " → " + nv
		}
		if len(reasons) > 0 || verdict != "" {
			out = append(out, Change{Advisory: d.Advisory, Reasons: reasons, Verdict: verdict})
		}
	}
	for _, d := range old.Derivations {
		if !seen[d.Advisory] {
			out = append(out, Change{Advisory: d.Advisory, Reasons: []string{"advisory no longer reported"}, Verdict: old.Verdicts[d.Advisory] + " → (none)"})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Advisory < out[j].Advisory })
	return out
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
