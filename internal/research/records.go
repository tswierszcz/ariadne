// Package research is the advisory-independent research interface
// (RESEARCH_INTERFACE RI1–RI3, docs/research-interface/CONTRACT.md,
// revision 1.0.0-draft.1): a frozen analysis bundle written by
// `ariadne research index` and the queries `ariadne research query`
// answers from it without rebuilding the graph.
//
// The wire types below follow docs/research-interface/v1/contract.schema.json.
// Struct fields are declared in alphabetical order of their JSON names, so
// encoding/json writes every record with sorted keys: identical analyses
// give identical bytes, which the bundle's hashes rely on.
package research

// Contract identifiers.
const (
	ContractRevision = "1.0.0-draft.1"
	BundleSchema     = "ariadne-research-bundle-1"
	IdentitySchema   = "ariadne-research-identity-1"
	QuerySchema      = "ariadne-research-query-1"
	ResponseSchema   = "ariadne-research-response-1"

	// Query rules this implementation applies (CONTRACT.md §8–§10).
	PathsRule     = "research-paths-1"
	RelevanceRule = "research-relevance-1"
	WorkRule      = "research-work-1"
	// DispatchRule marks client/supplied-dispatch edges: every edge leaving
	// a call site library mode classified as client code (LB2) or as a call
	// on a caller-supplied object.
	DispatchRule = "research-dispatch-1"
)

// Record types (the record_type discriminator).
const (
	TypeComponent = "component"
	TypeClass     = "class"
	TypeNode      = "node"
	TypeEdge      = "edge"
	TypeRoot      = "root"
	TypeGap       = "gap"
	TypeCall      = "call"
)

// Source is a method's source location hint.
type Source struct {
	Binding      string  `json:"binding"`
	LineEnd      *int    `json:"line_end"`
	LineStart    *int    `json:"line_start"`
	Path         *string `json:"path"`
	SourceSHA256 *string `json:"source_sha256"`
}

// Method is a method identity.
type Method struct {
	Class       string  `json:"class"`
	ClassSHA256 *string `json:"class_sha256"`
	Descriptor  string  `json:"descriptor"`
	InstanceID  *string `json:"instance_id"`
	MethodID    string  `json:"method_id"`
	Name        string  `json:"name"`
}

// Site is one call site or model anchor.
type Site struct {
	CallerNodeID string  `json:"caller_node_id"`
	Condition    *string `json:"condition"`
	Kind         string  `json:"kind"`
	PC           int     `json:"pc"`
}

// Component is a component instance.
type Component struct {
	ArtifactSHA256 string  `json:"artifact_sha256"`
	GAV            *string `json:"gav"`
	Identity       string  `json:"identity"`
	InstanceID     string  `json:"instance_id"`
	Location       string  `json:"location"`
	RecordType     string  `json:"record_type"`
	Role           string  `json:"role"`
}

// Class is one class definition, selected or shadowed.
type Class struct {
	Class           string `json:"class"`
	ClassID         string `json:"class_id"`
	ClassSHA256     string `json:"class_sha256"`
	InstanceID      string `json:"instance_id"`
	RecordType      string `json:"record_type"`
	Selected        bool   `json:"selected"`
	SelectedClassID string `json:"selected_class_id"`
	VariantJava     int    `json:"variant_java"`
}

// Node is a graph node.
type Node struct {
	AccessFlags     *int    `json:"access_flags"`
	Body            string  `json:"body"`
	CanonicalNodeID string  `json:"canonical_node_id"`
	Method          *Method `json:"method"`
	NodeID          string  `json:"node_id"`
	NodeKind        string  `json:"node_kind"`
	ReceiverType    *string `json:"receiver_type"`
	RecordType      string  `json:"record_type"`
	Source          Source  `json:"source"`
}

// Edge is a graph edge with every site that produced it.
type Edge struct {
	Basis            string `json:"basis"`
	Character        string `json:"character"`
	ClientDispatch   bool   `json:"client_dispatch"`
	EdgeID           string `json:"edge_id"`
	From             string `json:"from"`
	Kind             string `json:"kind"`
	RecordType       string `json:"record_type"`
	Rule             string `json:"rule"`
	Sites            []Site `json:"sites"`
	SuppliedDispatch bool   `json:"supplied_dispatch"`
	To               string `json:"to"`
}

// Root is an indexed root.
type Root struct {
	AnnouncedBy []string `json:"announced_by"`
	Basis       string   `json:"basis"`
	InheritedBy []string `json:"inherited_by"`
	NodeID      string   `json:"node_id"`
	RecordType  string   `json:"record_type"`
	RootID      string   `json:"root_id"`
}

// Gap is a recorded model limitation.
type Gap struct {
	CandidateInstanceIDs     []string `json:"candidate_instance_ids"`
	CandidateMethodIDs       []string `json:"candidate_method_ids"`
	GapID                    string   `json:"gap_id"`
	Mechanism                string   `json:"mechanism"`
	ReachableFromIndexedRoot bool     `json:"reachable_from_indexed_roots"`
	Reason                   string   `json:"reason"`
	RecordType               string   `json:"record_type"`
	Rule                     string   `json:"rule"`
	Site                     *Site    `json:"site"`
}

// Call is one bytecode call site with its declared operation.
type Call struct {
	BodyInInput        bool     `json:"body_in_input"`
	CallID             string   `json:"call_id"`
	DeclaredClass      string   `json:"declared_class"`
	DeclaredDescriptor string   `json:"declared_descriptor"`
	DeclaredName       string   `json:"declared_name"`
	RecordType         string   `json:"record_type"`
	Site               Site     `json:"site"`
	TargetNodeIDs      []string `json:"target_node_ids"`
}

// Provenance says where a document comes from.
type Provenance struct {
	Kind     string `json:"kind"`
	Producer string `json:"producer"`
	Revision string `json:"revision"`
}

// Manifest is manifest.json.
type Manifest struct {
	SchemaVersion    string `json:"schema_version"`
	ContractRevision string `json:"contract_revision"`
	AnalysisID       string `json:"analysis_id"`
	IdentityPath     string `json:"identity_path"`
	CreatedAt        string `json:"created_at"`
}

// Tool identifies the producer.
type Tool struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	BinarySHA256 *string           `json:"binary_sha256"`
	Algorithms   map[string]string `json:"algorithms"`
}

// Configuration is the effective static configuration.
type Configuration struct {
	RootsMode      string         `json:"roots_mode"`
	GraphMode      string         `json:"graph_mode"`
	TargetJava     int            `json:"target_java"`
	Profile        string         `json:"profile"`
	ActiveProfiles []string       `json:"active_profiles"`
	JPMSScenario   string         `json:"jpms_scenario"`
	APIExclude     []string       `json:"api_exclude"`
	Settings       map[string]any `json:"settings"`
}

// Input is one analyzed input file.
type Input struct {
	ID              string `json:"id"`
	Role            string `json:"role"`
	LogicalLocation string `json:"logical_location"`
	SHA256          string `json:"sha256"`
}

// Dataset is one JSONL file of the bundle.
type Dataset struct {
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	RecordCount int    `json:"record_count"`
	Encoding    string `json:"encoding"`
}

// FixedPoint is the model coordinator's outcome.
type FixedPoint struct {
	State   string   `json:"state"`
	Reasons []string `json:"reasons"`
}

// Coverage lists declared limitations.
type Coverage struct {
	Limits               []string `json:"limits"`
	DisabledParticipants []string `json:"disabled_participants"`
}

// Identity is the frozen analysis identity (identity.json).
type Identity struct {
	SchemaVersion    string        `json:"schema_version"`
	ContractRevision string        `json:"contract_revision"`
	Provenance       Provenance    `json:"provenance"`
	Tool             Tool          `json:"tool"`
	Configuration    Configuration `json:"configuration"`
	Inputs           []Input       `json:"inputs"`
	ClasspathOrder   []string      `json:"classpath_order"`
	Datasets         []Dataset     `json:"datasets"`
	Capabilities     []string      `json:"capabilities"`
	FixedPoint       FixedPoint    `json:"fixed_point"`
	Coverage         Coverage      `json:"coverage"`
}

// Capabilities this implementation answers.
var Capabilities = []string{"entries", "methods", "calls", "callers", "callees", "reachability"}

// Records is the content of a bundle's datasets.
type Records struct {
	Components []Component
	Classes    []Class
	Nodes      []Node
	Edges      []Edge
	Roots      []Root
	Gaps       []Gap
	Calls      []Call
}
