//go:build windows
// +build windows

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

	// Get Windows-specific file information
	if winStat, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		stat.Atime = time.Unix(0, winStat.LastAccessTime.Nanoseconds())
		stat.Mtime = time.Unix(0, winStat.LastWriteTime.Nanoseconds())
		stat.Ctime = time.Unix(0, winStat.CreationTime.Nanoseconds()) // Birth time on Windows
		return stat, true
	}

	return stat, false
}
