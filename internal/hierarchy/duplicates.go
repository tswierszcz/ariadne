package hierarchy

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
)

// SourceRef identifies one occurrence of a class on the classpath, for
// the Duplicate record below: which jar (or "" for the app, as
// ClassInfo.SourceJar already does) and a content hash of that exact
// .class entry, so two occurrences that merely share a name can still be
// told apart from two occurrences that are byte-for-byte the same class.
type SourceRef struct {
	Jar string
	// SHA256 is the hex-encoded content hash of this exact .class entry,
	// or "" if it could not be read back (the entry path is derived from
	// the FQCN by convention — Outer/Inner.class for Outer.Inner — which
	// holds for every jar ariadne has actually seen, but isn't a JVM
	// requirement, so a lookup failure is recorded as a missing hash, not
	// an error).
	SHA256 string
}

// Duplicate records one FQCN seen in more than one place on the
// classpath: Kept is the occurrence Build actually used (the first one
// in explicit job order — app, then dependency jars in the order
// ResolveDepJars/the CLI gave them), Shadowed is the one that lost.
// RECOMMENDATION JVM_GRAPH G1: "zachować listę duplikatów, ich hashe i
// przypisanie do artefaktów" — kept as structured data (not just the
// warning string) so a future consumer can tell a true shaded-jar
// collision (different SHA256) apart from the same class copy-pasted
// across two artifacts (identical SHA256).
type Duplicate struct {
	FQCN     string
	Kept     SourceRef
	Shadowed SourceRef
}

// hashClassEntry reads jarPath's class entry for fqcn and returns its
// SHA-256 as a hex string. ok is false if the entry couldn't be found
// at the conventional path or read — a missing hash, not a fatal error,
// since this is report/diagnostic evidence, not something correctness
// depends on.
func hashClassEntry(jarPath, fqcn string) (sum string, ok bool) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return "", false
	}
	defer zr.Close()

	entryName := strings.ReplaceAll(fqcn, ".", "/") + ".class"
	for _, f := range zr.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", false
		}
		defer rc.Close()
		h := sha256.New()
		if _, err := io.Copy(h, rc); err != nil {
			return "", false
		}
		return hex.EncodeToString(h.Sum(nil)), true
	}
	return "", false
}

// sourceRefOf builds a SourceRef for ci, lazily hashing its class entry.
func sourceRefOf(ci *ClassInfo) SourceRef {
	ref := SourceRef{Jar: ci.SourceJar}
	if sum, ok := hashClassEntry(ci.SourceJar, ci.FQCN); ok {
		ref.SHA256 = sum
	}
	return ref
}
