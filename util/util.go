package util

import (
	"encoding/base32"
	"errors"
	"fmt"
	"generator/constant"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	rng  *rand.Rand
	mu   sync.Mutex
	once sync.Once
)

// Seed initializes the deterministic PRNG with the provided seed.
// Call early (from main) to ensure reproducibility.
func Seed(seed int64) {
	mu.Lock()
	defer mu.Unlock()
	rng = rand.New(rand.NewSource(seed))
}

func GetAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	finfo, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if finfo.IsDir() {
		return absPath, nil
	}

	return filepath.Dir(absPath), nil
}

func GetRandomString(length int) string {
	// lazily initialize with a default seed if not seeded explicitly
	if rng == nil {
		once.Do(func() {
			mu.Lock()
			defer mu.Unlock()
			rng = rand.New(rand.NewSource(1))
		})
	}
	mu.Lock()
	defer mu.Unlock()
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = byte(rng.Intn(256))
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}

func GetFilePath(basedir string, ext string) string {
	filename := GetRandomString(constant.FileNameLen) + ext
	return filepath.Join(basedir, filename)
}

// SetTimes sets the modified and access times for a file or directory.
// On Unix systems, ctime cannot be set directly; this sets mtime/atime only.
func SetTimes(path string, atime, mtime time.Time) error {
	return os.Chtimes(path, atime, mtime)
}

// Touch updates or creates a file with provided content and times.
func Touch(path string, data []byte, atime, mtime time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, os.ModePerm); err != nil {
		return err
	}
	return SetTimes(path, atime, mtime)
}

// RemoveFile deletes a file and optionally sets directory times after deletion
// to emulate skew around deletion events. Skips silently if file doesn't exist.
func RemoveFile(path string, dirAtime, dirMtime *time.Time) error {
	// Check if file exists first
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, nothing to delete - skip silently
		return nil
	}

	if err := os.Remove(path); err != nil {
		return err
	}
	if dirAtime != nil && dirMtime != nil {
		_ = SetTimes(filepath.Dir(path), *dirAtime, *dirMtime)
	}
	return nil
}

// IsWindows reports whether the current OS is Windows.
func IsWindows() bool { return runtime.GOOS == "windows" }

// WriteADS writes data to an NTFS Alternate Data Stream for the given file path.
// The base file will be created if it does not exist. Windows-only.
func WriteADS(basePath, stream string, data []byte) error {
	if !IsWindows() {
		return errors.New("ADS not supported on non-Windows platforms")
	}
	if err := os.MkdirAll(filepath.Dir(basePath), os.ModePerm); err != nil {
		return err
	}
	// Ensure base file exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.WriteFile(basePath, []byte{}, os.ModePerm); err != nil {
			return err
		}
	}
	streamPath := fmt.Sprintf("%s:%s", basePath, stream)
	return os.WriteFile(streamPath, data, os.ModePerm)
}

// WriteMOTW writes a Mark-of-the-Web Zone.Identifier ADS with optional URLs.
// zoneID: 0 (My Computer), 1 (Local Intranet), 2 (Trusted), 3 (Internet), 4 (Restricted)
func WriteMOTW(basePath string, zoneID int, hostURL, referrerURL string) error {
	if !IsWindows() {
		return errors.New("MOTW not supported on non-Windows platforms")
	}
	content := "[ZoneTransfer]\r\n" + fmt.Sprintf("ZoneId=%d\r\n", zoneID)
	if referrerURL != "" {
		content += fmt.Sprintf("ReferrerUrl=%s\r\n", referrerURL)
	}
	if hostURL != "" {
		content += fmt.Sprintf("HostUrl=%s\r\n", hostURL)
	}
	return WriteADS(basePath, "Zone.Identifier", []byte(content))
}
