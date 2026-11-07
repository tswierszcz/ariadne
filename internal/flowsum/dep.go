// Package flowsum is DATA_FLOW D3's dependency summaries: for one method,
// which of its inputs — formal parameters, the receiver, fields, results
// of calls the analysis does not model — the value it returns, the
// values it writes to fields and the arguments it passes at each call
// site depend on. Where internal/dataflow resolves *values* (a literal,
// an allocation), this package tracks *dependence*: `return decode(p)`
// depends on parameter 0 even though its value is unknown. That is the
// shape a taint analysis (DATA_FLOW:D4) propagates.
//
// A summary is stated in the method's own terms (param:i, this, field:…,
// call:…) and applied separately at each call site by substituting that
// site's arguments, so two calls of one helper never mix their inputs
// (D3: "dwa call sites nie tworzą nieuzasadnionej ścieżki"). A call the
// analysis cannot resolve to exactly one analyzed method uses an
// explicit conservative policy — its result depends on every argument and
// the receiver, plus a `call:` source naming it — unless a catalog entry
// (catalog.go) states the exact dependence of a known JDK method. Control
// dependence (a branch on a value) is not tracked.
package flowsum

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// SummaryVersion versions the summary rules (domain, catalog, policies);
// a data snapshot records it, so a change invalidates a replay.
const SummaryVersion = "d3-1"

// Source kinds of a Dep element.
const (
	srcParam = "param:" // param:<descriptor index>
	srcThis  = "this"
	srcField = "field:" // field:<owner>.<name>
	srcCall  = "call:"  // call:<owner>#<name><desc> — an unmodeled call's result
	srcAlloc = "new@"   // new@<pc> — an object allocated in the method (internal)
)

// CallMany is the collapsed call source (see MaxCallSources).
const CallMany = srcCall + "*"

// MaxCallSources bounds the distinct unmodeled-call sources one Dep keeps;
// past it they collapse into CallMany ("results of many unmodeled
// calls"), which absorbs every other call source — a monotone top, so the
// fixed point still converges, and the dependence is never dropped.
const MaxCallSources = 32

// Sources are interned: a Dep is a sorted set of small integers, so the
// set operations of the fixed point compare integers, not strings.
type kind byte

const (
	kParam kind = iota + 1
	kThis
	kField
	kCall
	kCallMany
	kAlloc
)

var interner = struct {
	sync.RWMutex
	ids   map[string]int32
	names []string
	kinds []kind
	param []int // descriptor index for kParam, else -1
}{ids: map[string]int32{}}

func intern(s string) int32 {
	interner.RLock()
	id, ok := interner.ids[s]
	interner.RUnlock()
	if ok {
		return id
	}
	interner.Lock()
	defer interner.Unlock()
	if id, ok := interner.ids[s]; ok {
		return id
	}
	id = int32(len(interner.names))
	k, p := kField, -1
	switch {
	case s == srcThis:
		k = kThis
	case s == CallMany:
		k = kCallMany
	case strings.HasPrefix(s, srcParam):
		k = kParam
		p, _ = strconv.Atoi(s[len(srcParam):])
	case strings.HasPrefix(s, srcCall):
		k = kCall
	case strings.HasPrefix(s, srcAlloc):
		k = kAlloc
	}
	interner.ids[s] = id
	interner.names = append(interner.names, s)
	interner.kinds = append(interner.kinds, k)
	interner.param = append(interner.param, p)
	return id
}

func info(id int32) (string, kind, int) {
	interner.RLock()
	defer interner.RUnlock()
	return interner.names[id], interner.kinds[id], interner.param[id]
}

func kindOf(id int32) kind {
	interner.RLock()
	defer interner.RUnlock()
	return interner.kinds[id]
}

// Dep is a set of sources; nil is empty. A Dep is never modified after it
// is built.
type Dep []int32

func param(i int) string { return srcParam + strconv.Itoa(i) }

// subset reports whether every element of a is in b.
func subset(a, b Dep) bool {
	if len(a) > len(b) {
		return false
	}
	j := 0
	for _, x := range a {
		for j < len(b) && b[j] < x {
			j++
		}
		if j == len(b) || b[j] != x {
			return false
		}
		j++
	}
	return true
}

// union returns a ∪ b, allocating only when neither contains the other.
func union(a, b Dep) Dep {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 || subset(b, a) {
		return a
	}
	if subset(a, b) {
		return b
	}
	out := make(Dep, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, a[i])
			i++
			j++
		case a[i] < b[j]:
			out = append(out, a[i])
			i++
		default:
			out = append(out, b[j])
			j++
		}
	}
	out = append(out, a[i:]...)
	return capCalls(append(out, b[j:]...))
}

func capCalls(d Dep) Dep {
	if len(d) <= MaxCallSources {
		return d // cannot hold more call sources than elements
	}
	n, many := 0, false
	for _, id := range d {
		switch kindOf(id) {
		case kCall:
			n++
		case kCallMany:
			n++
			many = true
		}
	}
	if n <= MaxCallSources && (!many || n == 1) {
		return d
	}
	out := make(Dep, 0, len(d))
	for _, id := range d {
		if k := kindOf(id); k != kCall && k != kCallMany {
			out = append(out, id)
		}
	}
	return fromIDs(append(out, intern(CallMany)))
}

func fromIDs(ids []int32) Dep {
	if len(ids) == 0 {
		return nil
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := ids[:1]
	for _, id := range ids[1:] {
		if id != out[len(out)-1] {
			out = append(out, id)
		}
	}
	return out
}

func of(srcs ...string) Dep {
	ids := make([]int32, len(srcs))
	for i, s := range srcs {
		ids[i] = intern(s)
	}
	return fromIDs(ids)
}

func equal(a, b Dep) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (d Dep) has(s string) bool {
	interner.RLock()
	id, ok := interner.ids[s]
	interner.RUnlock()
	if !ok {
		return false
	}
	i := sort.Search(len(d), func(i int) bool { return d[i] >= id })
	return i < len(d) && d[i] == id
}

// Params returns the formal-parameter indexes d depends on, ascending.
func (d Dep) Params() []int {
	var out []int
	for _, id := range d {
		if _, k, p := info(id); k == kParam {
			out = append(out, p)
		}
	}
	sort.Ints(out)
	return out
}

// This reports whether d depends on the receiver.
func (d Dep) This() bool { return d.has(srcThis) }

// Fields returns the fields ("owner.name") d depends on, sorted.
func (d Dep) Fields() []string { return d.named(kField, srcField) }

// Calls returns the unmodeled calls ("owner#name+desc", or "*" for many)
// whose results d depends on, sorted.
func (d Dep) Calls() []string {
	return append(d.named(kCallMany, srcCall), d.named(kCall, srcCall)...)
}

func (d Dep) named(want kind, prefix string) []string {
	var out []string
	for _, id := range d {
		if s, k, _ := info(id); k == want {
			out = append(out, s[len(prefix):])
		}
	}
	sort.Strings(out)
	return out
}

// Strings is d's sources, sorted (for output).
func (d Dep) Strings() []string {
	out := make([]string, 0, len(d))
	for _, id := range d {
		s, _, _ := info(id)
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// String is the sources joined by spaces.
func (d Dep) String() string { return strings.Join(d.Strings(), " ") }

// substitute maps a callee's Dep into the caller: param:i → args[i],
// this → recv; fields and call sources are kept.
func substitute(d Dep, recv Dep, args []Dep) Dep {
	var out Dep
	var keep []int32
	for _, id := range d {
		_, k, p := info(id)
		switch k {
		case kThis:
			out = union(out, recv)
		case kParam:
			if p >= 0 && p < len(args) {
				out = union(out, args[p])
			}
		default:
			keep = append(keep, id)
		}
	}
	return union(out, fromIDs(keep))
}
