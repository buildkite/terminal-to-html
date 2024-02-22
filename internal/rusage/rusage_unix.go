//go:build unix

package rusage

import (
	"time"

	"golang.org/x/sys/unix"
)

// Stats returns the resources used by the program according to getrusage(2).
func Stats() (*Resources, error) {
	var usage unix.Rusage
	if err := unix.Getrusage(unix.RUSAGE_SELF, &usage); err != nil {
		return nil, err
	}

	return &Resources{
		Utime: time.Duration(usage.Utime.Nano()),
		Stime: time.Duration(usage.Stime.Nano()),

		// Note: These integer casts aren't redundant on 32-bit arches
		MaxRSS:      int64(usage.Maxrss),
		MinorFaults: int64(usage.Minflt),
		MajorFaults: int64(usage.Majflt),
		FSInBlocks:  int64(usage.Inblock),
		FSOutBlocks: int64(usage.Oublock),
	}, nil
}
