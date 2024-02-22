// Package rusage is a small wrapper around the POSIX system call getrusage.
package rusage

import "time"

// Resources summarises used system resources (mainly CPU time and memory).
// Some fields that appear in various rusage structs are omitted because we
// mainly care about Linux.
type Resources struct {
	// User-mode time and system-mode time
	Utime, Stime time.Duration

	// Maximum resident segment size, in platform-dependent units:
	// - Linux, Dragonfly, FreeBSD, NetBSD, OpenBSD, AIX: kilobytes
	// - Darwin: bytes
	// - Solaris, Illumos: pages
	MaxRSS int64

	// Counts of minor and major page faults
	MinorFaults, MajorFaults int64

	// Counts of file system performing input / output
	FSInBlocks, FSOutBlocks int64
}
