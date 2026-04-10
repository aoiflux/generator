package spec

// Manifest describes a set of operations to apply under a root directory.
type Manifest struct {
	Operations []Operation `yaml:"operations" json:"operations"`
}

// Operation represents a single action.
// Actions: create, update, append, delete, mace, rename, truncate, rotate, ads, motw
type Operation struct {
	Action     string `yaml:"action" json:"action"`
	Path       string `yaml:"path" json:"path"`               // required for most actions
	NewPath    string `yaml:"new_path" json:"new_path"`       // for rename/rotate
	Type       string `yaml:"type" json:"type"`               // for create: file|dir
	Ext        string `yaml:"ext" json:"ext"`                 // for create file when path is a directory
	Content    string `yaml:"content" json:"content"`         // optional literal content
	ContentLen int    `yaml:"content_len" json:"content_len"` // if Content empty, generate deterministic random of this length
	Atime      string `yaml:"atime" json:"atime"`             // RFC3339
	Mtime      string `yaml:"mtime" json:"mtime"`             // RFC3339
	// Windows-only extras
	Stream      string `yaml:"stream" json:"stream"`             // for ads: stream name (e.g., Zone.Identifier)
	ZoneID      int    `yaml:"zone_id" json:"zone_id"`           // for motw: ZoneId value (0-4)
	HostURL     string `yaml:"host_url" json:"host_url"`         // for motw
	ReferrerURL string `yaml:"referrer_url" json:"referrer_url"` // for motw
}

type Playbook struct {
	// Start time for the playbook, RFC3339 or "now"
	Start     string            `yaml:"start" json:"start"`
	Variables map[string]string `yaml:"variables" json:"variables"` // Global variables for templating
	Actors    []Actor           `yaml:"actors" json:"actors"`
	Steps     []Step            `yaml:"steps" json:"steps"`
}

type Actor struct {
	Name      string            `yaml:"name" json:"name"`
	Base      string            `yaml:"base" json:"base"`           // base directory under root for this actor
	Variables map[string]string `yaml:"variables" json:"variables"` // Actor-specific variables
}

type Step struct {
	Actor      string   `yaml:"actor" json:"actor"`
	Offset     string   `yaml:"offset" json:"offset"` // duration from playbook start for first iteration
	Every      string   `yaml:"every" json:"every"`   // repeat interval
	Repeat     int      `yaml:"repeat" json:"repeat"`
	Condition  string   `yaml:"condition" json:"condition"`     // Conditional execution: "odd", "even", "first", "last"
	BatchCount int      `yaml:"batch_count" json:"batch_count"` // Generate N files in this step
	Actions    []Action `yaml:"actions" json:"actions"`
}

// Action mirrors Operation but supports a relative time offset and template selection.
type Action struct {
	Action     string `yaml:"action" json:"action"`
	Path       string `yaml:"path" json:"path"`
	NewPath    string `yaml:"new_path" json:"new_path"`
	Type       string `yaml:"type" json:"type"`
	Ext        string `yaml:"ext" json:"ext"`
	Content    string `yaml:"content" json:"content"`
	ContentLen int    `yaml:"content_len" json:"content_len"`
	Template   string `yaml:"template" json:"template"`   // Predefined template: "email", "log", "script", "doc"
	Offset     string `yaml:"offset" json:"offset"`       // relative to step occurrence time
	Condition  string `yaml:"condition" json:"condition"` // Action-level condition
	// Optional explicit times override the computed time when provided
	Atime string `yaml:"atime" json:"atime"`
	Mtime string `yaml:"mtime" json:"mtime"`
	// Windows-only extras
	Stream      string `yaml:"stream" json:"stream"`
	ZoneID      int    `yaml:"zone_id" json:"zone_id"`
	HostURL     string `yaml:"host_url" json:"host_url"`
	ReferrerURL string `yaml:"referrer_url" json:"referrer_url"`
}
