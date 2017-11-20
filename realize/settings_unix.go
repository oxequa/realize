// +build !windows

package realize

import "syscall"

// Flimit defines the max number of watched files
func (s *Settings) Flimit() error {
	var rLimit syscall.Rlimit
	rLimit.Max = uint64(s.FileLimit)
	rLimit.Cur = uint64(s.FileLimit)

	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}
