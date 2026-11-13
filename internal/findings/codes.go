// Package findings holds the stable vocabulary of CVE:F0's results
// contract (docs/RESULTS_CONTRACT.MD): reason codes a finding can carry.
// CVE:F0 is a contract phase — nothing in the pipeline sets these codes
// yet; CVE:F1 attaches them to findings. The codes are part of the
// output contract once emitted, so they are never renamed or reused for a
// different meaning; a superseded code is deprecated, not deleted.
package findings

// ReasonCode is a stable, machine-readable reason attached to a finding.
type ReasonCode string

// Reason codes. Each one's meaning, axis and example are documented in
// docs/RESULTS_CONTRACT.MD ("Reason codes"); TestReasonCodesDocumented
// keeps the two lists identical.
const (
	// Component identity / version axis.
	VersionUnknown    ReasonCode = "VERSION_UNKNOWN"
	ClasspathConflict ReasonCode = "CLASSPATH_CONFLICT"
	// CVE:F6b: the instance's version is one of a bounded candidate set
	// its classes fit; the advisory affects at least one of them.
	VersionCandidates ReasonCode = "VERSION_CANDIDATES"

	// Mapping axis (which code the advisory concerns).
	MappingPartial        ReasonCode = "MAPPING_PARTIAL"
	MappingEntrypointOnly ReasonCode = "MAPPING_ENTRYPOINT_ONLY"
	MethodNotFound        ReasonCode = "METHOD_NOT_FOUND"
	SourceUnavailable     ReasonCode = "SOURCE_UNAVAILABLE"
	// ROADMAP B0: no source produced any mapping for the advisory, so no
	// verdict is possible (the UNKNOWN row says why).
	MappingMissing ReasonCode = "MAPPING_MISSING"
	// CVE:F4: versioned mappings.
	MappingVersionUncovered ReasonCode = "MAPPING_VERSION_UNCOVERED"
	MappingConflict         ReasonCode = "MAPPING_CONFLICT"
	// CVE:F5: a fix-diff mapping read on the post-fix side only.
	MappingNewSideOnly ReasonCode = "MAPPING_NEW_SIDE_ONLY"

	// Reachability-model axis.
	ModelIncomplete ReasonCode = "MODEL_INCOMPLETE"

	// Runtime axis.
	RuntimeUnsupported ReasonCode = "RUNTIME_UNSUPPORTED"
)

// All lists every defined reason code, in documentation order.
var All = []ReasonCode{
	VersionUnknown, ClasspathConflict, VersionCandidates,
	MappingPartial, MappingEntrypointOnly, MethodNotFound, SourceUnavailable, MappingMissing,
	MappingVersionUncovered, MappingConflict, MappingNewSideOnly,
	ModelIncomplete,
	RuntimeUnsupported,
}

// MappingQualifier reports a code that describes the mapping itself (its
// role, coverage or evidence) rather than a finding: v1 keeps it on the
// advisory row, v2 carries it in the mapping, never on an instance
// finding's reachability verdict.
func MappingQualifier(c ReasonCode) bool {
	return c == MappingPartial || c == MappingEntrypointOnly || c == MappingNewSideOnly
}
