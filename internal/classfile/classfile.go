package classfile

import (
	"fmt"
	"io"
)

const classMagic = 0xCAFEBABE

// Access flags (JVM spec 4.1, 4.5, 4.6) that callers care about; the rest
// are irrelevant to reachability analysis and are kept only as the raw
// uint16 on each struct.
const (
	AccPublic    = 0x0001
	AccStatic    = 0x0008
	AccFinal     = 0x0010
	AccInterface = 0x0200
	AccAbstract  = 0x0400
)

// Method is one method_info entry, resolved enough to build a call graph
// and detect profile entry points.
type Method struct {
	AccessFlags uint16
	Name        string
	Descriptor  string
	// Annotations are the FQCNs of this method's RuntimeVisibleAnnotations
	// (e.g. "org.springframework.web.bind.annotation.GetMapping").
	Annotations []string
	// ParameterAnnotations are the FQCNs of each parameter's own
	// RuntimeVisibleParameterAnnotations (e.g. CDI's
	// "javax.enterprise.event.Observes" on an observer method's event
	// parameter), nil if the method has no such attribute at all. Index i
	// holds parameter i's annotations; a parameter with none gets an empty
	// (not nil) slice, so len(ParameterAnnotations) still tells apart "no
	// attribute" from "attribute present, nothing annotated".
	ParameterAnnotations [][]string
	// Code is the raw bytecode from this method's Code attribute, or nil
	// for an abstract/native method.
	Code []byte
}

func (m Method) IsStatic() bool   { return m.AccessFlags&AccStatic != 0 }
func (m Method) IsPublic() bool   { return m.AccessFlags&AccPublic != 0 }
func (m Method) IsAbstract() bool { return m.AccessFlags&AccAbstract != 0 }

// Field is one field_info entry.
type Field struct {
	AccessFlags uint16
	Name        string
	Descriptor  string
}

// BootstrapMethod is one entry of the class-level BootstrapMethods
// attribute, with every MethodHandle argument already resolved (the only
// argument kind ariadne needs: the implementation-method handle a
// LambdaMetafactory bootstrap carries at argument index 1).
type BootstrapMethod struct {
	// Kind/Owner/Name/Desc/IsInterface describe the bootstrap method
	// itself (for LambdaMetafactory calls, always
	// java.lang.invoke.LambdaMetafactory#metafactory or #altMetafactory).
	Kind        byte
	Owner       string
	Name        string
	Desc        string
	IsInterface bool
	// Args holds one entry per bootstrap argument, in order; an argument
	// that isn't itself a MethodHandle (e.g. the samMethodType/
	// instantiatedMethodType MethodType arguments LambdaMetafactory also
	// takes) has IsMethodHandle == false and the other fields unset.
	Args []BootstrapArg
}

// BootstrapArg is one resolved bootstrap-method argument.
type BootstrapArg struct {
	IsMethodHandle bool
	Kind           byte
	Owner          string
	Name           string
	Desc           string
	IsInterface    bool
}

// IsLambdaMetafactory reports whether this bootstrap method is
// java.lang.invoke.LambdaMetafactory#metafactory or #altMetafactory —
// the only bootstrap ariadne knows how to resolve to a real target method.
func (b BootstrapMethod) IsLambdaMetafactory() bool {
	return b.Owner == "java.lang.invoke.LambdaMetafactory" &&
		(b.Name == "metafactory" || b.Name == "altMetafactory")
}

// ImplMethod returns the LambdaMetafactory bootstrap's implementation
// MethodHandle (its 2nd argument, 0-indexed 1), which is the method a
// lambda or method reference actually dispatches to at run time.
func (b BootstrapMethod) ImplMethod() (BootstrapArg, bool) {
	if !b.IsLambdaMetafactory() || len(b.Args) < 2 || !b.Args[1].IsMethodHandle {
		return BootstrapArg{}, false
	}
	return b.Args[1], true
}

// ClassFile is a parsed .class file, resolved enough for hierarchy
// analysis and call-graph extraction. Constant-pool indices are not
// exposed; everything callers need (class names, method refs, bootstrap
// targets) is already resolved to names during Parse.
type ClassFile struct {
	MinorVersion uint16
	MajorVersion uint16
	AccessFlags  uint16
	// ThisClass/SuperClass are FQCNs ("fixture.App", "java.lang.Object").
	// SuperClass is "" only for java.lang.Object itself.
	ThisClass  string
	SuperClass string
	// Interfaces are the FQCNs this class directly implements/extends.
	Interfaces  []string
	Fields      []Field
	Methods     []Method
	Annotations []string // class-level RuntimeVisibleAnnotations, as FQCNs
	Bootstrap   []BootstrapMethod

	pool pool // kept only so resolveInvokeDynamic (calls.go) can look up the CONSTANT_InvokeDynamic entry for each invokedynamic instruction
}

func (c ClassFile) IsInterface() bool { return c.AccessFlags&AccInterface != 0 }
func (c ClassFile) IsAbstract() bool  { return c.AccessFlags&AccAbstract != 0 }

// Parse reads one .class file. It returns an error that names what went
// wrong (a malformed or truncated file, or a constant-pool tag newer than
// this parser knows about) rather than panicking, so a caller walking many
// jars can skip one bad class and keep going.
func Parse(r io.Reader) (*ClassFile, error) {
	rd := newReader(r)

	magic := rd.u4()
	if rd.err == nil && magic != classMagic {
		return nil, fmt.Errorf("not a class file: bad magic %#x", magic)
	}

	minor := rd.u2()
	major := rd.u2()

	p, err := readPool(rd)
	if err != nil {
		return nil, err
	}
	if rd.err != nil {
		return nil, rd.err
	}

	accessFlags := rd.u2()
	thisIdx := rd.u2()
	superIdx := rd.u2()

	ifaceCount := rd.u2()
	interfaces := make([]string, 0, ifaceCount)
	for i := uint16(0); i < ifaceCount; i++ {
		idx := rd.u2()
		if rd.err != nil {
			break
		}
		name, err := p.className(idx)
		if err != nil {
			return nil, fmt.Errorf("interface %d: %w", i, err)
		}
		interfaces = append(interfaces, name)
	}

	fieldCount := rd.u2()
	fields := make([]Field, 0, fieldCount)
	for i := uint16(0); i < fieldCount && rd.err == nil; i++ {
		f, err := readField(rd, p)
		if err != nil {
			return nil, fmt.Errorf("field %d: %w", i, err)
		}
		fields = append(fields, f)
	}

	methodCount := rd.u2()
	methods := make([]Method, 0, methodCount)
	for i := uint16(0); i < methodCount && rd.err == nil; i++ {
		m, err := readMethod(rd, p)
		if err != nil {
			return nil, fmt.Errorf("method %d: %w", i, err)
		}
		methods = append(methods, m)
	}

	var classAnnotations []string
	var bootstrap []BootstrapMethod
	classAttrCount := rd.u2()
	for i := uint16(0); i < classAttrCount && rd.err == nil; i++ {
		name, data, err := readAttribute(rd, p)
		if err != nil {
			return nil, fmt.Errorf("class attribute %d: %w", i, err)
		}
		switch name {
		case "RuntimeVisibleAnnotations":
			classAnnotations, err = parseAnnotations(data, p)
			if err != nil {
				return nil, fmt.Errorf("class RuntimeVisibleAnnotations: %w", err)
			}
		case "BootstrapMethods":
			bootstrap, err = parseBootstrapMethods(data, p)
			if err != nil {
				return nil, fmt.Errorf("BootstrapMethods: %w", err)
			}
		}
	}

	if rd.err != nil {
		return nil, rd.err
	}

	thisName, err := p.className(thisIdx)
	if err != nil {
		return nil, fmt.Errorf("this_class: %w", err)
	}
	var superName string
	if superIdx != 0 {
		superName, err = p.className(superIdx)
		if err != nil {
			return nil, fmt.Errorf("super_class: %w", err)
		}
	}

	return &ClassFile{
		MinorVersion: minor,
		MajorVersion: major,
		AccessFlags:  accessFlags,
		ThisClass:    thisName,
		SuperClass:   superName,
		Interfaces:   interfaces,
		Fields:       fields,
		Methods:      methods,
		Annotations:  classAnnotations,
		Bootstrap:    bootstrap,
		pool:         p,
	}, nil
}

func readField(rd *reader, p pool) (Field, error) {
	accessFlags := rd.u2()
	nameIdx := rd.u2()
	descIdx := rd.u2()
	attrCount := rd.u2()
	for i := uint16(0); i < attrCount && rd.err == nil; i++ {
		if _, _, err := readAttribute(rd, p); err != nil {
			return Field{}, err
		}
	}
	if rd.err != nil {
		return Field{}, rd.err
	}
	name, err := p.utf8(nameIdx)
	if err != nil {
		return Field{}, fmt.Errorf("name: %w", err)
	}
	desc, err := p.utf8(descIdx)
	if err != nil {
		return Field{}, fmt.Errorf("descriptor: %w", err)
	}
	return Field{AccessFlags: accessFlags, Name: name, Descriptor: desc}, nil
}

func readMethod(rd *reader, p pool) (Method, error) {
	accessFlags := rd.u2()
	nameIdx := rd.u2()
	descIdx := rd.u2()
	attrCount := rd.u2()

	var annotations []string
	var paramAnnotations [][]string
	var code []byte
	for i := uint16(0); i < attrCount && rd.err == nil; i++ {
		name, data, err := readAttribute(rd, p)
		if err != nil {
			return Method{}, err
		}
		switch name {
		case "RuntimeVisibleAnnotations":
			annotations, err = parseAnnotations(data, p)
			if err != nil {
				return Method{}, fmt.Errorf("RuntimeVisibleAnnotations: %w", err)
			}
		case "RuntimeVisibleParameterAnnotations":
			paramAnnotations, err = parseParameterAnnotations(data, p)
			if err != nil {
				return Method{}, fmt.Errorf("RuntimeVisibleParameterAnnotations: %w", err)
			}
		case "Code":
			code, err = parseCodeAttribute(data)
			if err != nil {
				return Method{}, fmt.Errorf("Code: %w", err)
			}
		}
	}
	if rd.err != nil {
		return Method{}, rd.err
	}

	name, err := p.utf8(nameIdx)
	if err != nil {
		return Method{}, fmt.Errorf("name: %w", err)
	}
	desc, err := p.utf8(descIdx)
	if err != nil {
		return Method{}, fmt.Errorf("descriptor: %w", err)
	}
	return Method{
		AccessFlags:          accessFlags,
		Name:                 name,
		Descriptor:           desc,
		Annotations:          annotations,
		ParameterAnnotations: paramAnnotations,
		Code:                 code,
	}, nil
}

// readAttribute reads one attribute_info's name and raw bytes without
// interpreting them; callers that care about a particular attribute parse
// its data themselves.
func readAttribute(rd *reader, p pool) (name string, data []byte, err error) {
	nameIdx := rd.u2()
	length := rd.u4()
	data = rd.bytes(length)
	if rd.err != nil {
		return "", nil, rd.err
	}
	name, err = p.utf8(nameIdx)
	if err != nil {
		return "", nil, fmt.Errorf("attribute name: %w", err)
	}
	return name, data, nil
}
