//go:build !windows
// +build !windows

package timeline

import (
	"os"
	"syscall"
	"time"
)

type fileStat struct {
	Atime time.Time
	Mtime time.Time
	Ctime time.Time
}

func getFileTimes(info os.FileInfo) (fileStat, bool) {
	stat := fileStat{}

	// Get Unix-specific file information with proper timestamp extraction
	if sys := info.Sys(); sys != nil {
		if unixStat, ok := sys.(*syscall.Stat_t); ok {
			// Extract timestamps using platform-aware field access
			stat.Atime = timespecToTime(unixStat.Atim)
			stat.Mtime = timespecToTime(unixStat.Mtim)
			stat.Ctime = timespecToTime(unixStat.Ctim) // Change time on Unix (inode metadata change)
			return stat, true
		}
	}

	// Fallback: use ModTime for all if syscall extraction fails
	stat.Mtime = info.ModTime()
	stat.Atime = info.ModTime()
	stat.Ctime = info.ModTime()
	return stat, true
}

// timespecToTime converts syscall.Timespec to time.Time
func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(ts.Sec, ts.Nsec)
}
