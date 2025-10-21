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
	Start     string            `yaml:"start"`
	Variables map[string]string `yaml:"variables"` // Global variables for templating
	Actors    []Actor           `yaml:"actors"`
	Steps     []Step            `yaml:"steps"`
}

type Actor struct {
	Name      string            `yaml:"name"`
	Base      string            `yaml:"base"`      // base directory under root for this actor
	Variables map[string]string `yaml:"variables"` // Actor-specific variables
}

type Step struct {
	Actor      string   `yaml:"actor"`
	Offset     string   `yaml:"offset"` // duration from playbook start for first iteration
	Every      string   `yaml:"every"`  // repeat interval
	Repeat     int      `yaml:"repeat"`
	Condition  string   `yaml:"condition"`   // Conditional execution: "odd", "even", "first", "last"
	BatchCount int      `yaml:"batch_count"` // Generate N files in this step
	Actions    []Action `yaml:"actions"`
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
	Template   string `yaml:"template"`  // Predefined template: "email", "log", "script", "doc"
	Offset     string `yaml:"offset"`    // relative to step occurrence time
	Condition  string `yaml:"condition"` // Action-level condition
	// Optional explicit times override the computed time when provided
	Atime string `yaml:"atime"`
	Mtime string `yaml:"mtime"`
	// Windows-only extras (from manifest Operation)
	Stream      string `yaml:"stream"`
	ZoneID      int    `yaml:"zone_id"`
	HostURL     string `yaml:"host_url"`
	ReferrerURL string `yaml:"referrer_url"`
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
			// Check step-level condition
			if !evalCondition(st.Condition, i, st.Repeat) {
				continue
			}

			t := baseTime.Add(time.Duration(i) * everyDur)

			// Handle batch generation
			batchCount := st.BatchCount
			if batchCount <= 0 {
				batchCount = 1
			}

			for batchIdx := 0; batchIdx < batchCount; batchIdx++ {
				for _, a := range st.Actions {
					// Check action-level condition
					if !evalCondition(a.Condition, batchIdx, batchCount) {
						continue
					}

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

					// Templating context
					seq++
					ctx := templateContext{
						seq:       seq,
						batchIdx:  batchIdx,
						iteration: i,
						actor:     actor.Name,
						timestamp: at,
						variables: mergeVariables(p.Variables, actor.Variables),
					}

					path := renderTemplates(joinPath(actor.Base, a.Path), ctx)
					newPath := renderTemplates(joinPath(actor.Base, a.NewPath), ctx)

					// Handle content templating or templates
					content := ""
					if a.Template != "" {
						content = getTemplate(a.Template, ctx)
					} else {
						content = renderTemplates(a.Content, ctx)
					}

					ops = append(ops, manifestpkg.Operation{
						Action:      a.Action,
						Path:        path,
						NewPath:     newPath,
						Type:        a.Type,
						Ext:         a.Ext,
						Content:     content,
						ContentLen:  a.ContentLen,
						Atime:       atime,
						Mtime:       mtime,
						Stream:      a.Stream,
						ZoneID:      a.ZoneID,
						HostURL:     a.HostURL,
						ReferrerURL: a.ReferrerURL,
					})
				}
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
	reVar   = regexp.MustCompile(`\$\{VAR\:([^}]+)\}`)
	reUUID  = regexp.MustCompile(`\$\{UUID\}`)
	reIP    = regexp.MustCompile(`\$\{IP\}`)
	reHash  = regexp.MustCompile(`\$\{HASH\:(\d+)\}`)
	reBatch = regexp.MustCompile(`\$\{BATCH\}`)
	reIter  = regexp.MustCompile(`\$\{ITER\}`)
)

type templateContext struct {
	seq       int
	batchIdx  int
	iteration int
	actor     string
	timestamp time.Time
	variables map[string]string
}

// evalCondition evaluates conditional logic for steps/actions
func evalCondition(condition string, index, total int) bool {
	condition = strings.ToLower(strings.TrimSpace(condition))
	if condition == "" {
		return true
	}
	switch condition {
	case "odd":
		return index%2 == 1
	case "even":
		return index%2 == 0
	case "first":
		return index == 0
	case "last":
		return index == total-1
	default:
		return true
	}
}

// mergeVariables combines global and actor-specific variables
func mergeVariables(global, actor map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range global {
		result[k] = v
	}
	for k, v := range actor {
		result[k] = v
	}
	return result
}

// renderTemplates applies all template substitutions
func renderTemplates(s string, ctx templateContext) string {
	if s == "" {
		return s
	}
	out := s

	// ${RND:N} and ${RANDOM:N}
	out = reRnd.ReplaceAllStringFunc(out, func(m string) string {
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

	// ${SEQ}
	out = reSeq.ReplaceAllString(out, strconv.Itoa(ctx.seq))

	// ${BATCH}
	out = reBatch.ReplaceAllString(out, strconv.Itoa(ctx.batchIdx))

	// ${ITER}
	out = reIter.ReplaceAllString(out, strconv.Itoa(ctx.iteration))

	// ${DATE:layout}
	out = reDate.ReplaceAllStringFunc(out, func(m string) string {
		parts := reDate.FindStringSubmatch(m)
		if len(parts) != 2 {
			return m
		}
		return ctx.timestamp.Format(parts[1])
	})

	// ${ACTOR}
	if ctx.actor != "" {
		out = reActor.ReplaceAllString(out, ctx.actor)
	}

	// ${VAR:name}
	out = reVar.ReplaceAllStringFunc(out, func(m string) string {
		parts := reVar.FindStringSubmatch(m)
		if len(parts) != 2 {
			return m
		}
		if val, ok := ctx.variables[parts[1]]; ok {
			return val
		}
		return m
	})

	// ${UUID} - deterministic UUID based on seq
	out = reUUID.ReplaceAllString(out, fmt.Sprintf("%08x-0000-0000-0000-%012x", ctx.seq, ctx.seq))

	// ${IP} - deterministic IP based on seq
	out = reIP.ReplaceAllString(out, fmt.Sprintf("192.168.%d.%d", (ctx.seq/256)%256, ctx.seq%256))

	// ${HASH:N} - deterministic hash-like string of length N
	out = reHash.ReplaceAllStringFunc(out, func(m string) string {
		parts := reHash.FindStringSubmatch(m)
		if len(parts) != 2 {
			return m
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 {
			n = 32
		}
		return util.GetRandomString(n)
	})

	return out
}

// getTemplate returns predefined content templates
func getTemplate(templateName string, ctx templateContext) string {
	switch strings.ToLower(templateName) {
	case "email":
		return fmt.Sprintf(`From: %s@example.com
To: recipient@example.com
Subject: %s
Date: %s

This is an automated message from %s.

Message ID: %d
`,
			strings.ToLower(ctx.actor),
			util.GetRandomString(20),
			ctx.timestamp.Format(time.RFC1123Z),
			ctx.actor,
			ctx.seq)

	case "log":
		return fmt.Sprintf(`%s [INFO] User=%s Action=file_access File=document_%d.txt Result=success
%s [WARN] User=%s Action=failed_login Attempts=3
%s [INFO] User=%s Action=logout SessionID=%d
`,
			ctx.timestamp.Format(time.RFC3339),
			ctx.actor,
			ctx.seq,
			ctx.timestamp.Add(1*time.Minute).Format(time.RFC3339),
			ctx.actor,
			ctx.timestamp.Add(2*time.Minute).Format(time.RFC3339),
			ctx.actor,
			ctx.seq)

	case "script":
		return fmt.Sprintf(`#!/bin/bash
# Generated by %s at %s
# Sequence: %d

echo "Running automated task"
echo "User: %s"
echo "Timestamp: %s"
`,
			ctx.actor,
			ctx.timestamp.Format(time.RFC3339),
			ctx.seq,
			ctx.actor,
			ctx.timestamp.Format(time.RFC3339))

	case "doc":
		return fmt.Sprintf(`Document Title: Report %d
Author: %s
Date: %s

This is a generated document for testing purposes.
Document ID: %d
Generated content: %s
`,
			ctx.seq,
			ctx.actor,
			ctx.timestamp.Format("2006-01-02"),
			ctx.seq,
			util.GetRandomString(100))

	default:
		return util.GetRandomString(256)
	}
}

func renderTemplatesWithActor(s string, seq int, actor string) string {
	ctx := templateContext{
		seq:       seq,
		actor:     actor,
		timestamp: time.Now().UTC(),
	}
	return renderTemplates(s, ctx)
}

// writeTempManifest writes a small manifest to a temp file and returns its path
func writeTempManifest(ops []manifestpkg.Operation) string { return "" }
