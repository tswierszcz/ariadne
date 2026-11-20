package callback

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Listeners (ASYNC_CALLBACKS a3-5): an object registered with an event
// source runs when the source notifies — not when it is registered
// (ASYNC_CALLBACKS.MD: `addListener` alone does not create an event).
// Two event-source models:
//
//   - a listener list in a field: a store (add/offer/push/put… — a3-3's
//     queue store) into a JDK collection field, and a broadcast — a
//     method that walks that field's elements (iteration, get(i), peek)
//     without removing them or clearing the field, calling a method on
//     each. a3-3 pairs a store with a *drain* (elements removed or the
//     field cleared); a broadcast keeps them for the next event.
//   - java.beans.PropertyChangeSupport held in a field:
//     addPropertyChangeListener([name,] l) and firePropertyChange /
//     fireIndexedPropertyChange on the same field.
//
// The edge goes from the notifying call (the broadcast's call on the
// element, the fire call) to the method it runs on the listener — the
// class's resolution, a lambda's implementation — with condition
// "notified", and is gated (callgraph.NewCreatorGate) on the method where
// the listener enters: the registering method for a `new`/lambda, the
// caller passing it for a parameter (a3-2's forwarder). So a listener
// runs only once something both registers it and notifies; a listener
// registered only in code that never runs, or with a source that never
// notifies, gets nothing.
//
// Not modeled: listener arrays (copy-on-write `T[]` fields), a list
// walked through a parameter or a getter's result, forEach with a lambda,
// java.util.Observable, removal (a removed listener keeps its edge),
// and which property a named listener waits for.
const (
	// ListenersVersion names the rules; a snapshot records it.
	ListenersVersion = "a3-5"
	// ConditionNotified: the target runs when the source notifies its
	// registered listeners.
	ConditionNotified = "notified"
)

// ListenerStats summarizes the listener scan for the run log.
type ListenerStats struct {
	// Broadcasts: methods walking a listener field without draining it;
	// Registrations: stores into a broadcast field and
	// PropertyChangeSupport registrations; Fires: PropertyChangeSupport
	// fire calls on a field; Unresolved: registrations whose listener is
	// neither a `new`/lambda nor a parameter its callers fill with one.
	Broadcasts, Registrations, Fires, Unresolved int
}

const pcsOwner = "java.beans.PropertyChangeSupport"

type pcsFire struct {
	at    handoffMethod
	pc    int
	field string
	kind  classfile.CallKind
}

type pcsSites struct {
	adds  []queueStore
	fires []pcsFire
}

func isPCSAdd(c classfile.Call) bool {
	return c.Owner == pcsOwner && c.Name == "addPropertyChangeListener" &&
		(c.Desc == "(Ljava/beans/PropertyChangeListener;)V" || c.Desc == "(Ljava/lang/String;Ljava/beans/PropertyChangeListener;)V")
}

func pcsListenerArg(c classfile.Call) int {
	if strings.HasPrefix(c.Desc, "(Ljava/lang/String;") {
		return 1
	}
	return 0
}

func isPCSFire(c classfile.Call) bool {
	return c.Owner == pcsOwner && (c.Name == "firePropertyChange" || c.Name == "fireIndexedPropertyChange")
}

// anyField: v is read from a field (directly or through a static
// accessor, as collectionField); the field as "declaring class#name".
func anyField(h *hierarchy.Hierarchy, v valueflow.Value) (string, bool) {
	if v.Kind == valueflow.KindCallResult && v.Call != nil {
		v, _ = dataflow.ResolveCallResultOneHop(h, *v.Call)
	}
	if v.Kind != valueflow.KindField || v.Field == nil {
		return "", false
	}
	decl, ok := h.ResolveField(v.Field.Owner, v.Field.Name, v.Field.Desc)
	if !ok {
		return "", false
	}
	return decl + "#" + v.Field.Name, true
}

func listenerEdges(h *hierarchy.Hierarchy, fw *forwarder, stores []queueStore, bcasts []queueDrain, pcs pcsSites) ([]GatedEdge, ListenerStats) {
	var stats ListenerStats
	bcastMethods := map[string]bool{}
	for _, b := range bcasts {
		bcastMethods[b.at.key()] = true
	}
	stats.Broadcasts = len(bcastMethods)
	stats.Fires = len(pcs.fires)

	var out []GatedEdge
	idx := map[string]int{}
	add := func(from handoffMethod, pc int, kind classfile.CallKind, t handoffMethod, creator handoffMethod) {
		k := fmt.Sprintf("%s@%d>%s", from.key(), pc, t.key())
		c := MethodID{creator.fqcn, creator.name, creator.desc}
		if i, ok := idx[k]; ok {
			if !hasCreator(out[i].Creators, c) {
				out[i].Creators = append(out[i].Creators, c)
			}
			return
		}
		idx[k] = len(out)
		out = append(out, GatedEdge{CallerFQCN: from.fqcn, CallerName: from.name, CallerDesc: from.desc, PC: pc, Kind: kind,
			TargetFQCN: t.fqcn, TargetMethod: t.name, TargetMethodDesc: t.desc, Creators: []MethodID{c}, Condition: ConditionNotified})
	}

	// Listener lists in fields.
	for _, s := range stores {
		var matched []queueDrain
		for _, b := range bcasts {
			if b.at.key() != s.at.key() && pairs(h, s, b) {
				matched = append(matched, b)
			}
		}
		if len(matched) == 0 {
			continue
		}
		stats.Registrations++
		objs := queuedObjects(fw, s)
		if len(objs) == 0 {
			stats.Unresolved++
			continue
		}
		for _, o := range objs {
			for _, b := range matched {
				for _, c := range b.calls {
					if t, ok := runs(h, o.rc, c); ok {
						add(b.at, c.PC, c.Kind, t, o.from)
					}
				}
			}
		}
	}

	// PropertyChangeSupport.
	const pcl, pcDesc = "java.beans.PropertyChangeListener", "(Ljava/beans/PropertyChangeEvent;)V"
	call := classfile.Call{Kind: classfile.CallInterface, Owner: pcl, Name: "propertyChange", Desc: pcDesc}
	for _, s := range pcs.adds {
		var fires []pcsFire
		for _, f := range pcs.fires {
			if f.field == s.field {
				fires = append(fires, f)
			}
		}
		if len(fires) == 0 {
			continue
		}
		stats.Registrations++
		objs := queuedObjects(fw, s)
		if len(objs) == 0 {
			stats.Unresolved++
			continue
		}
		for _, o := range objs {
			t, ok := runsListener(h, o.rc, call)
			if !ok {
				continue
			}
			for _, f := range fires {
				add(f.at, f.pc, f.kind, t, o.from)
			}
		}
	}
	for i := range out {
		sort.Slice(out[i].Creators, func(a, b int) bool {
			x, y := out[i].Creators[a], out[i].Creators[b]
			return x.FQCN+"#"+x.Name+x.Desc < y.FQCN+"#"+y.Name+y.Desc
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		ka := a.CallerFQCN + "#" + a.CallerName + a.CallerDesc + fmt.Sprintf("@%06d", a.PC) + a.TargetFQCN + "#" + a.TargetMethod + a.TargetMethodDesc
		kb := b.CallerFQCN + "#" + b.CallerName + b.CallerDesc + fmt.Sprintf("@%06d", b.PC) + b.TargetFQCN + "#" + b.TargetMethod + b.TargetMethodDesc
		return ka < kb
	})
	return out, stats
}

// runsListener: runs for a JDK listener interface — the class's own
// resolution (no subtype check: the hierarchy does not hold JDK types'
// subtype edges for every class), or a lambda of that method.
func runsListener(h *hierarchy.Hierarchy, rc ResolvedCallback, c classfile.Call) (handoffMethod, bool) {
	switch {
	case rc.TargetFQCN != "":
		decl, ok := h.ResolveExact(rc.TargetFQCN, c.Name, c.Desc)
		if !ok {
			return handoffMethod{}, false
		}
		return handoffMethod{decl, c.Name, c.Desc}, true
	case rc.ImplOwner != "":
		if rc.SamMethod != c.Name {
			return handoffMethod{}, false
		}
		return handoffMethod{rc.ImplOwner, rc.ImplName, rc.ImplDesc}, true
	}
	return handoffMethod{}, false
}
