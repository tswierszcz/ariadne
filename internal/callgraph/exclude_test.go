package callgraph

import (
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// TestExcludeAbstractBodies (builder 5): with an Exclude, a body an
// abstract class declares is a CHA target only while some non-excluded
// concrete subtype selects it; a body no concrete subtype selects keeps
// CHA's usual treatment; without an Exclude nothing changes.
func TestExcludeAbstractBodies(t *testing.T) {
	const d = "()V"
	m := func(name string, abstract bool) hierarchy.Method {
		mm := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: d}}
		if abstract {
			mm.AccessFlags = classfile.AccAbstract
		}
		return mm
	}
	cls := func(fqcn, super string, abstract bool, methods ...hierarchy.Method) *hierarchy.ClassInfo {
		ci := &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, Methods: methods}
		if abstract {
			ci.AccessFlags = classfile.AccAbstract
		}
		return ci
	}
	caller := &hierarchy.ClassInfo{FQCN: "p.Caller", SuperClass: "java.lang.Object", Methods: []hierarchy.Method{{
		Method: classfile.Method{Name: "run", Descriptor: d},
		Calls:  []classfile.Call{{Kind: classfile.CallVirtual, Owner: "p.Base", Name: "append", Desc: d, PC: 1}},
	}}}
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		caller,
		cls("p.Base", "java.lang.Object", true, m("append", true)),
		// SMTPAppenderBase's shape: its only concrete subtype is excluded.
		cls("p.MailBase", "p.Base", true, m("append", false)),
		cls("p.Mail", "p.MailBase", false),
		// Inherited by an excluded and a kept subtype: stays.
		cls("p.FileBase", "p.Base", true, m("append", false)),
		cls("p.RollingFile", "p.FileBase", false),
		cls("p.File", "p.FileBase", false),
		// Overridden by its only concrete subtype: that subtype is the
		// target, the abstract body is not selected by anyone.
		cls("p.NetBase", "p.Base", true, m("append", false)),
		cls("p.Net", "p.NetBase", false, m("append", false)),
		// No concrete subtype in the input: CHA's usual treatment.
		cls("p.Orphan", "p.Base", true, m("append", false)),
		// log4j's WriterAppender/ConsoleAppender shape: the declaring
		// class is excluded, a kept subtype inherits its body.
		cls("p.Writer", "p.Base", false, m("append", false)),
		cls("p.Console", "p.Writer", false),
	})
	excluded := map[string]bool{"p.Mail": true, "p.RollingFile": true, "p.Writer": true}
	targets := func(s Strategy) map[string]bool {
		out := map[string]bool{}
		for _, t := range s.DispatchTargets(h, caller.Methods[0].Calls[0]) {
			out[t] = true
		}
		return out
	}
	with := targets(CHA{Exclude: func(c string) bool { return excluded[c] }, ExcludeRule: "test"})
	for target, want := range map[string]bool{"p.MailBase": false, "p.FileBase": true, "p.Net": true, "p.Orphan": true, "p.Writer": true} {
		if with[target] != want {
			t.Errorf("with Exclude: %s target = %v, want %v (targets %v)", target, with[target], want, with)
		}
	}
	without := targets(CHA{})
	for _, target := range []string{"p.MailBase", "p.FileBase", "p.NetBase", "p.Net", "p.Orphan"} {
		if !without[target] {
			t.Errorf("without Exclude: %s missing (targets %v)", target, without)
		}
	}
}
