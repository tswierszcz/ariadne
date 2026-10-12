package classfile

import "fmt"

// readPool reads the constant_pool_count and the constant pool itself.
// constant_pool_count is the number of entries plus one (entry 0 is
// unused), and a Long/Double entry takes its own index plus an unusable
// placeholder right after it, per the JVM spec's historical quirk.
func readPool(rd *reader) (pool, error) {
	count := rd.u2()
	if rd.err != nil {
		return nil, rd.err
	}
	p := make(pool, count)

	for i := uint16(1); i < count; i++ {
		tag := rd.u1()
		if rd.err != nil {
			return nil, rd.err
		}
		e := cpEntry{tag: tag}

		switch tag {
		case tagUtf8:
			n := rd.u2()
			b := rd.bytes(uint32(n))
			e.str = string(b)
		case tagInteger, tagFloat:
			rd.u4()
		case tagLong, tagDouble:
			rd.u4()
			rd.u4()
		case tagClass, tagString, tagMethodType, tagModule, tagPackage:
			e.nameIndex = rd.u2()
		case tagFieldref, tagMethodref, tagInterfaceMethodref:
			e.classIndex = rd.u2()
			e.nameAndTypeIndex = rd.u2()
		case tagNameAndType:
			e.nameIndex = rd.u2()
			e.descriptorIndex = rd.u2()
		case tagMethodHandle:
			e.referenceKind = rd.u1()
			e.referenceIndex = rd.u2()
		case tagDynamic, tagInvokeDynamic:
			e.bootstrapMethodAttrIndex = rd.u2()
			e.nameAndTypeIndex = rd.u2()
		default:
			return nil, fmt.Errorf("constant pool entry %d: unknown tag %d (class file from a Java version newer than this parser knows about)", i, tag)
		}
		if rd.err != nil {
			return nil, rd.err
		}
		p[i] = e

		if tag == tagLong || tag == tagDouble {
			i++ // the next index is an unusable placeholder; skip it
		}
	}
	return p, nil
}
