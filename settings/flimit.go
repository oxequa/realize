// +build !windows

package settings

import "syscall"

// Flimit defines the max number of watched files
func (s *Settings) Flimit() {
	var rLimit syscall.Rlimit
	rLimit.Max = s.Config.Flimit
	rLimit.Cur = s.Config.Flimit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		s.Fatal("Error Setting Rlimit", err)
	}
}
