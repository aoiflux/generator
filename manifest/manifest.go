package manifest

import (
	"fmt"
	"generator/util"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Manifest describes a set of operations to apply under a root directory.
type Manifest struct {
	Operations []Operation `yaml:"operations"`
}

// Operation represents a single action.
// Actions: create, update, delete, mace, rename
type Operation struct {
	Action     string `yaml:"action"`
	Path       string `yaml:"path"`        // required for all except when action is none
	NewPath    string `yaml:"new_path"`    // for rename
	Type       string `yaml:"type"`        // for create: file|dir
	Ext        string `yaml:"ext"`         // for create file when path is a directory
	Content    string `yaml:"content"`     // optional literal content
	ContentLen int    `yaml:"content_len"` // if Content empty, generate deterministic random of this length
	Atime      string `yaml:"atime"`       // RFC3339; for mace/update/create/delete(dir times)
	Mtime      string `yaml:"mtime"`       // RFC3339
	// Windows-only extras
	Stream      string `yaml:"stream"`       // for ads: stream name (e.g., Zone.Identifier)
	ZoneID      int    `yaml:"zone_id"`      // for motw: ZoneId value (0-4)
	HostURL     string `yaml:"host_url"`     // for motw
	ReferrerURL string `yaml:"referrer_url"` // for motw
}

// ExecuteManifest reads a YAML manifest file and applies operations under root.
func ExecuteManifest(root string, manifestPath string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}
	return ExecuteOperations(root, m.operationsSequential())
}

func (m Manifest) operationsSequential() []Operation {
	// already sequential; in future may support groups
	return m.Operations
}

// ExecuteOperations applies the provided operations under the given root, sequentially.
func ExecuteOperations(root string, ops []Operation) error {
	for idx, op := range ops {
		if err := executeOp(root, op); err != nil {
			return fmt.Errorf("op %d (%s %s): %w", idx+1, op.Action, op.Path, err)
		}
	}
	return nil
}

func executeOp(root string, op Operation) error {
	action := strings.ToLower(strings.TrimSpace(op.Action))
	target := filepath.Join(root, filepath.FromSlash(op.Path))

	switch action {
	case "create":
		if strings.ToLower(op.Type) == "dir" || (op.Ext == "" && strings.HasSuffix(op.Path, string(os.PathSeparator))) {
			return os.MkdirAll(target, fs.ModePerm)
		}
		// file create
		content := []byte(op.Content)
		if len(content) == 0 {
			sz := op.ContentLen
			if sz <= 0 {
				sz = 1024
			}
			content = []byte(util.GetRandomString(sz))
		}
		if op.Ext != "" && filepath.Ext(target) == "" {
			target = target + op.Ext
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, os.ModePerm); err != nil {
			return err
		}
		at, mt, _ := parseTimes(op.Atime, op.Mtime)
		if !at.IsZero() || !mt.IsZero() {
			if at.IsZero() {
				at = mt
			}
			if mt.IsZero() {
				mt = at
			}
			_ = util.SetTimes(target, at, mt)
		}
		return nil

	case "update":
		content := []byte(op.Content)
		if len(content) == 0 {
			sz := op.ContentLen
			if sz <= 0 {
				sz = 1024
			}
			content = []byte(util.GetRandomString(sz))
		}
		if err := os.WriteFile(target, content, os.ModePerm); err != nil {
			return err
		}
		at, mt, _ := parseTimes(op.Atime, op.Mtime)
		if !at.IsZero() || !mt.IsZero() {
			if at.IsZero() {
				at = mt
			}
			if mt.IsZero() {
				mt = at
			}
			_ = util.SetTimes(target, at, mt)
		}
		return nil

	case "append":
		content := []byte(op.Content)
		if len(content) == 0 {
			sz := op.ContentLen
			if sz <= 0 {
				sz = 256
			}
			content = []byte(util.GetRandomString(sz))
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			return err
		}
		if _, err := f.Write(content); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		at, mt, _ := parseTimes(op.Atime, op.Mtime)
		if !at.IsZero() || !mt.IsZero() {
			if at.IsZero() {
				at = mt
			}
			if mt.IsZero() {
				mt = at
			}
			_ = util.SetTimes(target, at, mt)
		}
		return nil

	case "delete":
		var dirAt, dirMt *time.Time
		if at, mt, ok := parseTimes(op.Atime, op.Mtime); ok {
			dirAt, dirMt = &at, &mt
		}
		return util.RemoveFile(target, dirAt, dirMt)

	case "mace":
		at, mt, _ := parseTimes(op.Atime, op.Mtime)
		if at.IsZero() && mt.IsZero() {
			return fmt.Errorf("mace requires at least one of atime/mtime")
		}
		if at.IsZero() {
			at = mt
		}
		if mt.IsZero() {
			mt = at
		}
		return util.SetTimes(target, at, mt)

	case "rename":
		if strings.TrimSpace(op.NewPath) == "" {
			return fmt.Errorf("rename requires new_path")
		}
		newTarget := filepath.Join(root, filepath.FromSlash(op.NewPath))
		if err := os.MkdirAll(filepath.Dir(newTarget), os.ModePerm); err != nil {
			return err
		}
		return os.Rename(target, newTarget)

	case "truncate":
		// Truncate file to zero bytes (create if missing)
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		at, mt, _ := parseTimes(op.Atime, op.Mtime)
		if !at.IsZero() || !mt.IsZero() {
			if at.IsZero() {
				at = mt
			}
			if mt.IsZero() {
				mt = at
			}
			_ = util.SetTimes(target, at, mt)
		}
		return nil

	case "rotate":
		// Rotate: rename original to new_path, then create empty file at original path
		if strings.TrimSpace(op.NewPath) == "" {
			return fmt.Errorf("rotate requires new_path")
		}
		newTarget := filepath.Join(root, filepath.FromSlash(op.NewPath))
		if err := os.MkdirAll(filepath.Dir(newTarget), os.ModePerm); err != nil {
			return err
		}
		if err := os.Rename(target, newTarget); err != nil {
			return err
		}
		// recreate empty original
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		// Optionally set times on both files
		if at, mt, ok := parseTimes(op.Atime, op.Mtime); ok {
			_ = util.SetTimes(newTarget, at, mt)
			_ = util.SetTimes(target, at, mt)
		}
		return nil

	case "ads":
		if !util.IsWindows() {
			return fmt.Errorf("ads action is only supported on Windows")
		}
		if strings.TrimSpace(op.Stream) == "" {
			return fmt.Errorf("ads requires stream name")
		}
		content := []byte(op.Content)
		if len(content) == 0 {
			sz := op.ContentLen
			if sz <= 0 {
				sz = 128
			}
			content = []byte(util.GetRandomString(sz))
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		return util.WriteADS(target, op.Stream, content)

	case "motw":
		if !util.IsWindows() {
			return fmt.Errorf("motw action is only supported on Windows")
		}
		zone := op.ZoneID
		if zone < 0 || zone > 4 {
			zone = 3
		}
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}
		return util.WriteMOTW(target, zone, op.HostURL, op.ReferrerURL)

	default:
		return fmt.Errorf("unknown action: %s", op.Action)
	}
}

func parseTimes(at, mt string) (time.Time, time.Time, bool) {
	var atT, mtT time.Time
	ok := false
	if strings.TrimSpace(at) != "" {
		if t, err := time.Parse(time.RFC3339, at); err == nil {
			atT = t
			ok = true
		}
	}
	if strings.TrimSpace(mt) != "" {
		if t, err := time.Parse(time.RFC3339, mt); err == nil {
			mtT = t
			ok = true
		}
	}
	return atT, mtT, ok
}
