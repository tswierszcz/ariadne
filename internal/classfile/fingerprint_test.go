package classfile

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"ariadne/internal/relocate"
)

// appendUtf8 returns b with one unused Utf8 constant appended to the
// constant pool: every byte after the pool moves, every index stays —
// a minimal stand-in for a shading tool re-serializing a class.
func appendUtf8(t *testing.T, b []byte, s string) []byte {
	t.Helper()
	count := int(binary.BigEndian.Uint16(b[8:10]))
	pos := 10
	for i := 1; i < count; i++ {
		tag := b[pos]
		pos++
		switch tag {
		case tagUtf8:
			n := int(binary.BigEndian.Uint16(b[pos:]))
			pos += 2 + n
		case tagInteger, tagFloat, tagFieldref, tagMethodref, tagInterfaceMethodref, tagNameAndType, tagDynamic, tagInvokeDynamic:
			pos += 4
		case tagLong, tagDouble:
			pos += 8
			i++
		case tagClass, tagString, tagMethodType, tagModule, tagPackage:
			pos += 2
		case tagMethodHandle:
			pos += 3
		default:
			t.Fatalf("unknown tag %d", tag)
		}
	}
	var out bytes.Buffer
	out.Write(b[:8])
	binary.Write(&out, binary.BigEndian, uint16(count+1))
	out.Write(b[10:pos])
	out.WriteByte(tagUtf8)
	binary.Write(&out, binary.BigEndian, uint16(len(s)))
	out.WriteString(s)
	out.Write(b[pos:])
	return out.Bytes()
}

func parseBytes(t *testing.T, b []byte) *ClassFile {
	t.Helper()
	cf, err := Parse(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return cf
}

// TestFingerprint: a re-serialized class (different bytes, same content)
// keeps its fingerprint; a different class does not share it.
func TestFingerprint(t *testing.T) {
	codec, err := os.ReadFile("../../testdata/jpms/classes/auto/fixture/jpms/auto/AutoCodec.class")
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := os.ReadFile("../../testdata/jpms/classes/auto/fixture/jpms/auto/AutoHasher.class")
	if err != nil {
		t.Fatal(err)
	}
	padded := appendUtf8(t, codec, "padding")
	if bytes.Equal(codec, padded) {
		t.Fatal("padding changed nothing")
	}
	a, err := Fingerprint(parseBytes(t, codec))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Fingerprint(parseBytes(t, padded))
	c, _ := Fingerprint(parseBytes(t, hasher))
	if a != b {
		t.Errorf("re-serialized class has another fingerprint")
	}
	if a == c {
		t.Errorf("two different classes share a fingerprint")
	}

	// Any change of code, here one constant operand, changes it.
	cf := parseBytes(t, codec)
	for i := range cf.Methods {
		for j, op := range cf.Methods[i].Code {
			if op == 0x2a { // aload_0 → aload_1: same length, other code
				cf.Methods[i].Code[j] = 0x2b
				d, _ := Fingerprint(cf)
				if d == a {
					t.Errorf("changed code kept the fingerprint")
				}
				return
			}
		}
	}
	t.Fatal("no aload_0 in the fixture")
}

// TestCodeFingerprintNormalizesWideForms: ldc_w and ldc of the same
// constant, and a branch whose offset shifts with an instruction's width,
// fingerprint alike.
func TestCodeFingerprintNormalizesWideForms(t *testing.T) {
	cf := &ClassFile{pool: pool{{}, {tag: tagUtf8, str: "x"}, {tag: tagString, nameIndex: 1}}}
	// ldc #2; ifnull → return (pc 2 + 3)
	short := []byte{0x12, 0x02, 0xc6, 0x00, 0x03, 0xb1}
	// ldc_w #2; ifnull → return (pc 3 + 3): another offset, same target
	wide := []byte{0x13, 0x00, 0x02, 0xc6, 0x00, 0x03, 0xb1}
	a, _ := cf.codeFingerprint(short, relocate.Set{})
	b, _ := cf.codeFingerprint(wide, relocate.Set{})
	if a != b {
		t.Errorf("ldc vs ldc_w:\n%s\n%s", a, b)
	}
	other := []byte{0x12, 0x02, 0xc6, 0x00, 0x03, 0xb0} // areturn instead of return
	if c, _ := cf.codeFingerprint(other, relocate.Set{}); c == a {
		t.Errorf("different code, same fingerprint")
	}
}
