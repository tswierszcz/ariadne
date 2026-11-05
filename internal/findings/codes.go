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

	// Mapping axis (which code the advisory concerns).
	MappingPartial        ReasonCode = "MAPPING_PARTIAL"
	MappingEntrypointOnly ReasonCode = "MAPPING_ENTRYPOINT_ONLY"
	MethodNotFound        ReasonCode = "METHOD_NOT_FOUND"
	SourceUnavailable     ReasonCode = "SOURCE_UNAVAILABLE"

	// Reachability-model axis.
	ModelIncomplete ReasonCode = "MODEL_INCOMPLETE"

	// Runtime axis.
	RuntimeUnsupported ReasonCode = "RUNTIME_UNSUPPORTED"
)

// All lists every defined reason code, in documentation order.
var All = []ReasonCode{
	VersionUnknown, ClasspathConflict,
	MappingPartial, MappingEntrypointOnly, MethodNotFound, SourceUnavailable,
	ModelIncomplete,
	RuntimeUnsupported,
}
