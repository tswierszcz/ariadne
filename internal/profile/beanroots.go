package profile

import (
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

// Spring bean lifecycle roots (Roots.Beans; ROADMAP B4 b4-7).
//
// The Spring container creates a bean and calls into it through
// reflection: BeanUtils#instantiateClass → Constructor#newInstance for the
// constructor, Method#invoke for @Autowired/@Inject/@Resource/@Value
// methods, for @PostConstruct/@PreDestroy methods and for a @Bean's
// initMethod/destroyMethod (an inferred close()/shutdown() too). None of
// these sites has a literal the reflection models resolve, so a real-JVM
// oracle (testdata/di) measured every such body unreachable from the
// profile's roots: the constructor of a @Service, its @PostConstruct, an
// @Autowired setter. Everything they call was invisible under -roots
// profile. The callback interfaces (InitializingBean, the Aware
// interfaces, FactoryBean, BeanPostProcessor, ApplicationListener, …) are
// not listed: Spring calls them through those interfaces, and CHA already
// reaches the application's implementations (the oracle measured them
// reached).
//
// A bean is an application class carrying a stereotype (StereotypeBeans'
// list), @Configuration or @ControllerAdvice/@RestControllerAdvice, or a
// @Bean method's product (for its init/destroy methods only — the factory
// method itself is a root of the spring profile). A bean whose class-level
// @Profile is known to be inactive (-active-profiles given) is left out.
// Anything else the container decides at run time — a @Conditional, @Lazy
// with no injection point, a @Profile without -active-profiles — counts as
// "may be created": the rule never guesses false.

// BeanRootKind says why a member is a lifecycle root.
type BeanRootKind string

const (
	BeanRootConstructor BeanRootKind = "constructor"
	BeanRootInjection   BeanRootKind = "injection method"
	BeanRootInit        BeanRootKind = "init annotation"
	BeanRootDestroy     BeanRootKind = "destroy annotation"
	BeanRootInitMethod  BeanRootKind = "@Bean initMethod"
	BeanRootDestroyMeth BeanRootKind = "@Bean destroyMethod"
)

// BeanRoot is one lifecycle method the container may call: FQCN#Name+Desc
// is the declaring class's body.
type BeanRoot struct {
	Bean, FQCN, Name, Desc string
	Kind                   BeanRootKind
}

var lifecycleBeanAnnotations = map[string]bool{
	configurationFQCN: true,
	"org.springframework.web.bind.annotation.ControllerAdvice":     true,
	"org.springframework.web.bind.annotation.RestControllerAdvice": true,
}

var injectionAnnotations = []string{
	"org.springframework.beans.factory.annotation.Autowired",
	"org.springframework.beans.factory.annotation.Value",
	"javax.inject.Inject",
	"jakarta.inject.Inject",
	"javax.annotation.Resource",
	"jakarta.annotation.Resource",
}

var initAnnotations = []string{"javax.annotation.PostConstruct", "jakarta.annotation.PostConstruct"}

var destroyAnnotations = []string{"javax.annotation.PreDestroy", "jakarta.annotation.PreDestroy"}

// constructorAnnotations mark the constructor Spring picks among several.
var constructorAnnotations = []string{
	"org.springframework.beans.factory.annotation.Autowired",
	"javax.inject.Inject",
	"jakarta.inject.Inject",
}

// BeanLifecycleRoots lists the lifecycle methods of every bean in h (see
// the comment above). activeProfiles nil means no -active-profiles: every
// @Profile is unknown and its bean counts.
func BeanLifecycleRoots(h *hierarchy.Hierarchy, activeProfiles []string) []BeanRoot {
	var active map[string]bool
	if activeProfiles != nil {
		active = make(map[string]bool, len(activeProfiles))
		for _, p := range activeProfiles {
			active[p] = true
		}
	}
	var out []BeanRoot
	seen := map[string]bool{}
	add := func(r BeanRoot) {
		key := r.FQCN + "#" + r.Name + r.Desc
		if !seen[key] {
			seen[key] = true
			out = append(out, r)
		}
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp || ci.IsInterface() || ci.AccessFlags&classfile.AccAbstract != 0 {
			continue
		}
		bean := false
		for _, a := range ci.Annotations {
			if stereotypeAnnotations[a] || lifecycleBeanAnnotations[a] {
				bean = true
				break
			}
		}
		if !bean {
			continue
		}
		if pc, ok := profileCondition(ci.AnnotationDetails, active, fqcn, ""); ok && pc.Active == ConditionFalse {
			continue
		}
		if c, ok := chosenConstructor(ci); ok {
			add(BeanRoot{Bean: fqcn, FQCN: fqcn, Name: c.Name, Desc: c.Descriptor, Kind: BeanRootConstructor})
		}
		for _, r := range annotatedLifecycleMethods(h, fqcn) {
			add(r)
		}
	}
	for _, r := range beanMethodLifecycle(h, active) {
		add(r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].FQCN != out[j].FQCN {
			return out[i].FQCN < out[j].FQCN
		}
		return out[i].Name+out[i].Desc < out[j].Name+out[j].Desc
	})
	return out
}

// chosenConstructor is the constructor Spring instantiates a bean with:
// the only one declared; else the one carrying @Autowired/@Inject (the
// first, if several); else the no-argument one. Several unannotated
// constructors without a no-argument one: none (Spring fails).
func chosenConstructor(ci *hierarchy.ClassInfo) (hierarchy.Method, bool) {
	ctors := ci.MethodsByName("<init>")
	if len(ctors) == 1 {
		return ctors[0], true
	}
	for _, c := range ctors {
		if hasAny(c.Annotations, constructorAnnotations) {
			return c, true
		}
	}
	for _, c := range ctors {
		if c.Descriptor == "()V" {
			return c, true
		}
	}
	return hierarchy.Method{}, false
}

// annotatedLifecycleMethods walks the bean class and its application
// superclasses (Spring's injection and lifecycle metadata covers the whole
// class hierarchy) for non-static injection, init and destroy methods.
func annotatedLifecycleMethods(h *hierarchy.Hierarchy, bean string) []BeanRoot {
	var out []BeanRoot
	for cur := bean; cur != ""; {
		ci, ok := h.Lookup(cur)
		if !ok || !ci.IsApp {
			break
		}
		for _, m := range ci.Methods {
			if m.IsStatic() || m.IsAbstract() || m.Name == "<init>" || m.Name == "<clinit>" {
				continue
			}
			kind := BeanRootKind("")
			switch {
			case hasAny(m.Annotations, initAnnotations):
				kind = BeanRootInit
			case hasAny(m.Annotations, destroyAnnotations):
				kind = BeanRootDestroy
			case hasAny(m.Annotations, injectionAnnotations):
				kind = BeanRootInjection
			}
			if kind != "" {
				out = append(out, BeanRoot{Bean: bean, FQCN: cur, Name: m.Name, Desc: m.Descriptor, Kind: kind})
			}
		}
		cur = ci.SuperClass
	}
	return out
}

// inferredDestroy is Spring's "(inferred)" destroyMethod, the @Bean
// default: a public no-argument close, else shutdown.
var inferredDestroy = []string{"close", "shutdown"}

// beanMethodLifecycle reads @Bean(initMethod, destroyMethod) on the
// @Configuration classes' factory methods. The product's class is the
// declared return type when concrete, plus every subtype of it the factory
// method creates with `new` (a method declared to return DataSource that
// builds a HikariDataSource); the init/destroy method is resolved on each
// by name with no parameters, as Spring does. With destroyMethod unset,
// Spring infers close() or shutdown(); destroyMethod="" turns that off.
func beanMethodLifecycle(h *hierarchy.Hierarchy, active map[string]bool) []BeanRoot {
	var out []BeanRoot
	for _, fqcn := range h.SortedFQCNs() {
		ci, ok := h.Lookup(fqcn)
		if !ok || !ci.IsApp || !hasAny(ci.Annotations, []string{configurationFQCN}) {
			continue
		}
		if pc, ok := profileCondition(ci.AnnotationDetails, active, fqcn, ""); ok && pc.Active == ConditionFalse {
			continue
		}
		for _, m := range ci.Methods {
			ann, found := findAnnotation(m.AnnotationDetails, beanMethodFQCN)
			if !found {
				continue
			}
			if pc, ok := profileCondition(m.AnnotationDetails, active, fqcn, m.Name); ok && pc.Active == ConditionFalse {
				continue
			}
			ret, ok := methodReturnTypeFQCN(m.Descriptor)
			if !ok {
				continue
			}
			products := productClasses(h, ret, m)
			if name, ok := stringElement(ann, "initMethod"); ok && name != "" {
				for _, p := range products {
					out = append(out, noArgMethods(h, p, []string{name}, BeanRootInitMethod)...)
				}
			}
			destroy, set := stringElement(ann, "destroyMethod")
			names := []string{destroy}
			if !set || destroy == "(inferred)" {
				names = inferredDestroy
			} else if destroy == "" {
				names = nil
			}
			for _, p := range products {
				out = append(out, noArgMethods(h, p, names, BeanRootDestroyMeth)...)
			}
		}
	}
	return out
}

func stringElement(ann classfile.Annotation, name string) (string, bool) {
	v, ok := ann.Value(name)
	if !ok {
		return "", false
	}
	s, ok := v.Const.(string)
	return s, ok
}

// productClasses: the concrete classes a @Bean method may return.
func productClasses(h *hierarchy.Hierarchy, ret string, factory hierarchy.Method) []string {
	var out []string
	seen := map[string]bool{}
	concrete := func(fqcn string) bool {
		ci, ok := h.Lookup(fqcn)
		return ok && !ci.IsInterface() && ci.AccessFlags&classfile.AccAbstract == 0
	}
	if concrete(ret) {
		seen[ret] = true
		out = append(out, ret)
	}
	subtypes := map[string]bool{}
	for _, s := range h.Subtypes(ret) {
		subtypes[s] = true
	}
	for _, r := range factory.InitRefs {
		if r.Kind == classfile.InitNew && subtypes[r.Owner] && !seen[r.Owner] && concrete(r.Owner) {
			seen[r.Owner] = true
			out = append(out, r.Owner)
		}
	}
	sort.Strings(out)
	return out
}

// noArgMethods resolves, for each name in order, the first public or
// declared no-argument method of that name on class or its superclasses;
// for an inferred list (several names) only the first found counts, as in
// Spring.
func noArgMethods(h *hierarchy.Hierarchy, class string, names []string, kind BeanRootKind) []BeanRoot {
	for _, name := range names {
		for cur := class; cur != ""; {
			ci, ok := h.Lookup(cur)
			if !ok {
				break
			}
			for _, m := range ci.MethodsByName(name) {
				if strings.HasPrefix(m.Descriptor, "()") && !m.IsStatic() && !m.IsAbstract() {
					return []BeanRoot{{Bean: class, FQCN: cur, Name: m.Name, Desc: m.Descriptor, Kind: kind}}
				}
			}
			cur = ci.SuperClass
		}
	}
	return nil
}

// BeanLifecycleAppRoots is BeanLifecycleRoots' call-graph-node form.
func BeanLifecycleAppRoots(g *callgraph.Graph, roots []BeanRoot) []uint32 {
	var out []uint32
	for _, r := range roots {
		if id, ok := g.ID(r.FQCN, r.Name, r.Desc); ok {
			out = append(out, id)
		}
	}
	return out
}
