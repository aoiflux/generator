//go:build !windows
// +build !windows

package timeline

import (
	"os"
	"time"
)

type fileStat struct {
	Atime time.Time
	Mtime time.Time
	Ctime time.Time
}

func getFileTimes(info os.FileInfo) (fileStat, bool) {
	stat := fileStat{}

	// On Unix systems, we'll use the ModTime and fall back for other times
	// since accessing syscall.Stat_t fields is platform-specific (Linux vs BSD/macOS)
	stat.Mtime = info.ModTime()
	stat.Atime = info.ModTime() // Fallback: use mtime as atime
	stat.Ctime = info.ModTime() // Fallback: use mtime as ctime

	return stat, true
}
