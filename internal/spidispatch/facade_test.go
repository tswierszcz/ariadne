package spidispatch

import (
	"reflect"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/models"
)

const (
	connectDesc = "(Ljava/lang/String;Ljava/util/Properties;)Ljava/sql/Connection;"
	getConnDesc = "(Ljava/lang/String;)Ljava/sql/Connection;"
)

func method(name, desc string, calls ...classfile.Call) hierarchy.Method {
	return hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls}
}

// facadeInput: an app calling DriverManager.getConnection (static) and a
// look-alike instance getConnection on its own type; a driver in its own
// jar implementing Driver#connect, and its connection implementing
// java.sql.Connection.
func facadeInput() *hierarchy.Hierarchy {
	return hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: "app.Main", SuperClass: "java.lang.Object", SourceJar: "app.jar", IsApp: true, Methods: []hierarchy.Method{
			method("direct", "()V",
				classfile.Call{Kind: classfile.CallStatic, Owner: "java.sql.DriverManager", Name: "getConnection", Desc: getConnDesc, PC: 3},
				classfile.Call{Kind: classfile.CallInterface, Owner: "java.sql.Connection", Name: "close", Desc: "()V", PC: 9}),
			method("lookAlike", "()V",
				classfile.Call{Kind: classfile.CallVirtual, Owner: "java.sql.DriverManager", Name: "getConnection", Desc: getConnDesc, PC: 4}),
		}},
		{FQCN: "drv.Driver", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.Driver"}, SourceJar: "drv.jar", Methods: []hierarchy.Method{
			method("connect", connectDesc),
		}},
		{FQCN: "drv.Conn", SuperClass: "java.lang.Object", Interfaces: []string{"java.sql.Connection"}, SourceJar: "drv.jar", Methods: []hierarchy.Method{
			method("close", "()V"),
		}},
	})
}

// TestDriverManagerFacade (p2): a static DriverManager#getConnection call
// reaches Driver#connect of every registered driver, and makes the
// driver's connection a product; without a registration, or for a call of
// another kind, nothing is modeled.
func TestDriverManagerFacade(t *testing.T) {
	h := facadeInput()
	reg := []Registration{{Service: "java.sql.Driver", Provider: "drv.Driver", Source: "drv.jar"}}
	c := Find(h, reg, callgraph.DefaultSkip)
	getConn := models.Signature{Owner: "java.sql.DriverManager", Name: "getConnection", Desc: getConnDesc}
	closeSig := models.Signature{Owner: "java.sql.Connection", Name: "close", Desc: "()V"}
	if !reflect.DeepEqual(c.Signatures, []models.Signature{closeSig, getConn}) {
		t.Fatalf("signatures = %v", c.Signatures)
	}
	if c.Products["java.sql.Connection"] != "java.sql.Driver" {
		t.Errorf("products = %v, want java.sql.Connection from java.sql.Driver", c.Products)
	}
	ev := NewEvaluator(h, c)
	m := &models.Model{ID: ModelID}
	caller := models.Signature{Owner: "app.Main", Name: "direct", Desc: "()V"}
	fact := ev.Evaluate(m, models.CallSite{Caller: caller, Call: getConn, PC: 3})
	want := []models.Target{{Method: models.Signature{Owner: "drv.Driver", Name: "connect", Desc: connectDesc}, Kind: classfile.CallInterface}}
	if !reflect.DeepEqual(fact.Targets, want) {
		t.Errorf("facade targets = %v, want %v", fact.Targets, want)
	}
	fact = ev.Evaluate(m, models.CallSite{Caller: caller, Call: closeSig, PC: 9})
	if len(fact.Targets) != 1 || fact.Targets[0].Method.Owner != "drv.Conn" {
		t.Errorf("product targets = %v, want drv.Conn#close", fact.Targets)
	}
	// Not static: not the facade.
	lookAlike := models.CallSite{Caller: models.Signature{Owner: "app.Main", Name: "lookAlike", Desc: "()V"}, Call: getConn, PC: 4}
	if fact := ev.Evaluate(m, lookAlike); len(fact.Targets) != 0 {
		t.Errorf("non-static getConnection targets = %v, want none", fact.Targets)
	}

	// No registered driver: the facade is not a signature.
	if c := Find(h, nil, callgraph.DefaultSkip); len(c.Signatures) != 0 {
		t.Errorf("signatures without a registration = %v", c.Signatures)
	}
}
