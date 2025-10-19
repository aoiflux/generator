package playbook

import (
	"fmt"
	manifestpkg "generator/manifest"
	"generator/util"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Playbook struct {
	// Start time for the playbook, RFC3339 or "now"
	Start  string  `yaml:"start"`
	Actors []Actor `yaml:"actors"`
	Steps  []Step  `yaml:"steps"`
}

type Actor struct {
	Name string `yaml:"name"`
	Base string `yaml:"base"` // base directory under root for this actor
}

type Step struct {
	Actor   string   `yaml:"actor"`
	Offset  string   `yaml:"offset"` // duration from playbook start for first iteration
	Every   string   `yaml:"every"`  // repeat interval
	Repeat  int      `yaml:"repeat"`
	Actions []Action `yaml:"actions"`
}

// Action mirrors manifest Operation but supports a relative time offset and templating in fields
type Action struct {
	Action     string `yaml:"action"`
	Path       string `yaml:"path"`
	NewPath    string `yaml:"new_path"`
	Type       string `yaml:"type"`
	Ext        string `yaml:"ext"`
	Content    string `yaml:"content"`
	ContentLen int    `yaml:"content_len"`
	Offset     string `yaml:"offset"` // relative to step occurrence time
	// Optional explicit times override the computed time when provided
	Atime string `yaml:"atime"`
	Mtime string `yaml:"mtime"`
}

// ExecutePlaybook loads a playbook YAML and executes compiled operations under root.
func ExecutePlaybook(root string, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pb Playbook
	if err := yaml.Unmarshal(data, &pb); err != nil {
		return fmt.Errorf("parse playbook: %w", err)
	}
	ops, err := pb.compile(root)
	if err != nil {
		return err
	}
	// Execute sequentially to preserve deterministic order
	if err := manifestpkg.ExecuteOperations(root, ops); err != nil {
		return err
	}
	return nil
}

// compile playbook into concrete manifest operations with resolved times and paths
func (p Playbook) compile(root string) ([]manifestpkg.Operation, error) {
	startTime := time.Now().UTC()
	if strings.TrimSpace(p.Start) != "" && strings.ToLower(p.Start) != "now" {
		t, err := time.Parse(time.RFC3339, p.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid start: %w", err)
		}
		startTime = t
	}
	actors := map[string]Actor{}
	for _, a := range p.Actors {
		actors[strings.ToLower(a.Name)] = a
	}

	var ops []manifestpkg.Operation
	seq := 0
	for _, st := range p.Steps {
		actor, ok := actors[strings.ToLower(st.Actor)]
		if !ok {
			return nil, fmt.Errorf("unknown actor: %s", st.Actor)
		}
		offsetDur, err := parseDurationSafe(st.Offset)
		if err != nil {
			return nil, fmt.Errorf("step offset: %w", err)
		}
		everyDur, err := parseDurationDefault(st.Every, 0)
		if err != nil {
			return nil, fmt.Errorf("step every: %w", err)
		}
		if st.Repeat <= 0 {
			st.Repeat = 1
		}
		baseTime := startTime.Add(offsetDur)

		for i := 0; i < st.Repeat; i++ {
			t := baseTime.Add(time.Duration(i) * everyDur)
			for _, a := range st.Actions {
				at := t
				if d, err := parseDurationSafe(a.Offset); err == nil {
					at = t.Add(d)
				}

				// Determine final timestamps
				atime := a.Atime
				mtime := a.Mtime
				if strings.TrimSpace(atime) == "" {
					atime = at.Format(time.RFC3339)
				}
				if strings.TrimSpace(mtime) == "" {
					mtime = at.Format(time.RFC3339)
				}

				// Templating on fields
				seq++
				path := renderTemplatesWithActor(joinPath(actor.Base, a.Path), seq, actor.Name)
				newPath := renderTemplatesWithActor(joinPath(actor.Base, a.NewPath), seq, actor.Name)
				content := renderTemplatesWithActor(a.Content, seq, actor.Name)

				ops = append(ops, manifestpkg.Operation{
					Action:     a.Action,
					Path:       path,
					NewPath:    newPath,
					Type:       a.Type,
					Ext:        a.Ext,
					Content:    content,
					ContentLen: a.ContentLen,
					Atime:      atime,
					Mtime:      mtime,
				})
			}
		}
	}

	return ops, nil
}

func joinPath(base, p string) string {
	if strings.TrimSpace(p) == "" {
		return p
	}
	if strings.TrimSpace(base) == "" {
		return p
	}
	// keep forward slash YAML style, convert later by manifest executor
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(p, "/")
}

func parseDurationDefault(s string, def time.Duration) (time.Duration, error) {
	if strings.TrimSpace(s) == "" {
		return def, nil
	}
	return time.ParseDuration(s)
}

func parseDurationSafe(s string) (time.Duration, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

var (
	reRnd   = regexp.MustCompile(`\$\{(RND|RANDOM)\:(\d+)\}`)
	reSeq   = regexp.MustCompile(`\$\{SEQ\}`)
	reDate  = regexp.MustCompile(`\$\{DATE\:([^}]+)\}`)
	reActor = regexp.MustCompile(`\$\{ACTOR\}`)
)

func renderTemplatesWithActor(s string, seq int, actor string) string {
	if s == "" {
		return s
	}
	out := reRnd.ReplaceAllStringFunc(s, func(m string) string {
		parts := reRnd.FindStringSubmatch(m)
		if len(parts) != 3 {
			return m
		}
		n, err := strconv.Atoi(parts[2])
		if err != nil || n <= 0 {
			n = 8
		}
		return util.GetRandomString(n)
	})
	out = reSeq.ReplaceAllString(out, strconv.Itoa(seq))
	now := time.Now().UTC()
	out = reDate.ReplaceAllStringFunc(out, func(m string) string {
		parts := reDate.FindStringSubmatch(m)
		if len(parts) != 2 {
			return m
		}
		return now.Format(parts[1])
	})
	if actor != "" {
		out = reActor.ReplaceAllString(out, actor)
	}
	return out
}

// writeTempManifest writes a small manifest to a temp file and returns its path
func writeTempManifest(ops []manifestpkg.Operation) string { return "" }
