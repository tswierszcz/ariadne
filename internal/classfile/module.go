package classfile

import (
	"bytes"
	"fmt"
)

// AccModule is the class-level ACC_MODULE flag (JVM spec 4.1): set only
// on a module-info.class, which declares a module rather than a class.
const AccModule = 0x8000

// Module flags (JVM spec 4.7.25) shared by module_flags and requires_flags.
const (
	ModuleAccOpen       = 0x0020 // module_flags only: an open module
	ModuleAccTransitive = 0x0020 // requires_flags only
	ModuleAccStaticPh   = 0x0040 // requires_flags only: requires static
	ModuleAccSynthetic  = 0x1000
	ModuleAccMandated   = 0x8000
)

// ModuleInfo is a decoded Module attribute (JVM spec 4.7.25) — the
// declarations a module-info.class carries (SERVICE_LOADING S3). It is a
// declaration, not an execution fact: `provides S with P` says which
// provider a ServiceLoader *on the module path* would see, never that
// the provider is constructed, and on the class path the JVM ignores the
// whole descriptor.
//
// Class names (Uses, Provides) are FQCNs in dot form, same as
// ClassFile.ThisClass; package names (Exports/Opens) are also converted
// to dot form ("com.example.api"); module names are kept as written.
type ModuleInfo struct {
	Name     string
	Flags    uint16
	Version  string // "" when module_version_index is 0
	Requires []ModuleRequires
	Exports  []ModulePackages
	Opens    []ModulePackages
	Uses     []string
	Provides []ModuleProvides
}

// IsOpen reports whether this is an `open module`.
func (m ModuleInfo) IsOpen() bool { return m.Flags&ModuleAccOpen != 0 }

// ModuleRequires is one requires entry.
type ModuleRequires struct {
	Module  string
	Flags   uint16
	Version string // compiled-against version, "" when absent
}

// ModulePackages is one exports or opens entry; To is empty for an
// unqualified export/open.
type ModulePackages struct {
	Package string
	Flags   uint16
	To      []string
}

// ModuleProvides is one `provides Service with Impl1, Impl2` entry, in
// declaration order (ServiceLoader's own module-path iteration order).
type ModuleProvides struct {
	Service string
	With    []string
}

// IsModule reports whether this class file is a module descriptor.
func (c ClassFile) IsModule() bool { return c.AccessFlags&AccModule != 0 }

// moduleName resolves a CONSTANT_Module entry to its name.
func (p pool) moduleName(idx uint16) (string, error) {
	e, err := p.get(idx)
	if err != nil {
		return "", err
	}
	if e.tag != tagModule {
		return "", fmt.Errorf("constant pool index %d is tag %d, want Module", idx, e.tag)
	}
	return p.utf8(e.nameIndex)
}

// packageName resolves a CONSTANT_Package entry to its dot-form name.
func (p pool) packageName(idx uint16) (string, error) {
	e, err := p.get(idx)
	if err != nil {
		return "", err
	}
	if e.tag != tagPackage {
		return "", fmt.Errorf("constant pool index %d is tag %d, want Package", idx, e.tag)
	}
	name, err := p.utf8(e.nameIndex)
	if err != nil {
		return "", err
	}
	return binaryNameToFQCN(name), nil
}

// optionalUtf8 resolves a Utf8 index that the spec allows to be 0.
func (p pool) optionalUtf8(idx uint16) (string, error) {
	if idx == 0 {
		return "", nil
	}
	return p.utf8(idx)
}

func parseModuleAttribute(data []byte, p pool) (*ModuleInfo, error) {
	rd := newReader(bytes.NewReader(data))
	nameIdx := rd.u2()
	flags := rd.u2()
	versionIdx := rd.u2()
	if rd.err != nil {
		return nil, rd.err
	}
	name, err := p.moduleName(nameIdx)
	if err != nil {
		return nil, fmt.Errorf("module name: %w", err)
	}
	version, err := p.optionalUtf8(versionIdx)
	if err != nil {
		return nil, fmt.Errorf("module version: %w", err)
	}
	m := &ModuleInfo{Name: name, Flags: flags, Version: version}

	reqCount := rd.u2()
	for i := uint16(0); i < reqCount && rd.err == nil; i++ {
		modIdx, rflags, verIdx := rd.u2(), rd.u2(), rd.u2()
		if rd.err != nil {
			break
		}
		mod, err := p.moduleName(modIdx)
		if err != nil {
			return nil, fmt.Errorf("requires %d: %w", i, err)
		}
		ver, err := p.optionalUtf8(verIdx)
		if err != nil {
			return nil, fmt.Errorf("requires %d version: %w", i, err)
		}
		m.Requires = append(m.Requires, ModuleRequires{Module: mod, Flags: rflags, Version: ver})
	}

	readPackages := func(kind string) ([]ModulePackages, error) {
		count := rd.u2()
		var out []ModulePackages
		for i := uint16(0); i < count && rd.err == nil; i++ {
			pkgIdx, pflags, toCount := rd.u2(), rd.u2(), rd.u2()
			toIdx := make([]uint16, 0, toCount)
			for j := uint16(0); j < toCount && rd.err == nil; j++ {
				toIdx = append(toIdx, rd.u2())
			}
			if rd.err != nil {
				break
			}
			pkg, err := p.packageName(pkgIdx)
			if err != nil {
				return nil, fmt.Errorf("%s %d: %w", kind, i, err)
			}
			entry := ModulePackages{Package: pkg, Flags: pflags}
			for _, idx := range toIdx {
				to, err := p.moduleName(idx)
				if err != nil {
					return nil, fmt.Errorf("%s %d target: %w", kind, i, err)
				}
				entry.To = append(entry.To, to)
			}
			out = append(out, entry)
		}
		return out, nil
	}
	if m.Exports, err = readPackages("exports"); err != nil {
		return nil, err
	}
	if m.Opens, err = readPackages("opens"); err != nil {
		return nil, err
	}

	usesCount := rd.u2()
	for i := uint16(0); i < usesCount && rd.err == nil; i++ {
		idx := rd.u2()
		if rd.err != nil {
			break
		}
		svc, err := p.className(idx)
		if err != nil {
			return nil, fmt.Errorf("uses %d: %w", i, err)
		}
		m.Uses = append(m.Uses, svc)
	}

	provCount := rd.u2()
	for i := uint16(0); i < provCount && rd.err == nil; i++ {
		svcIdx, withCount := rd.u2(), rd.u2()
		withIdx := make([]uint16, 0, withCount)
		for j := uint16(0); j < withCount && rd.err == nil; j++ {
			withIdx = append(withIdx, rd.u2())
		}
		if rd.err != nil {
			break
		}
		svc, err := p.className(svcIdx)
		if err != nil {
			return nil, fmt.Errorf("provides %d: %w", i, err)
		}
		entry := ModuleProvides{Service: svc}
		for _, idx := range withIdx {
			impl, err := p.className(idx)
			if err != nil {
				return nil, fmt.Errorf("provides %d with: %w", i, err)
			}
			entry.With = append(entry.With, impl)
		}
		m.Provides = append(m.Provides, entry)
	}
	if rd.err != nil {
		return nil, fmt.Errorf("truncated Module attribute: %w", rd.err)
	}
	return m, nil
}
