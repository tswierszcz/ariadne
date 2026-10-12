package classfile

import "bytes"

// parseCodeAttribute parses a Code attribute's bytes (JVM spec 4.7.3) and
// returns just the raw bytecode array; the exception table and nested
// attributes (LineNumberTable, StackMapTable, ...) are skipped, since call
// extraction only needs the instructions themselves.
func parseCodeAttribute(data []byte) ([]byte, error) {
	rd := newReader(bytes.NewReader(data))
	rd.u2() // max_stack
	rd.u2() // max_locals
	codeLen := rd.u4()
	code := rd.bytes(codeLen)

	excCount := rd.u2()
	rd.skip(uint32(excCount) * 8) // 4 u2 fields per exception_table entry

	attrCount := rd.u2()
	for i := uint16(0); i < attrCount && rd.err == nil; i++ {
		rd.u2() // attribute_name_index (unresolved: we don't care which attribute this is)
		length := rd.u4()
		rd.skip(length)
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return code, nil
}
