// +build windows

package realize

import "syscall"

// isHidden check if a file or a path is hidden
func isHidden(path string) bool {
	p, e := syscall.UTF16PtrFromString(path)
	if e != nil {
		return false
	}
	attrs, e := syscall.GetFileAttributes(p)
	if e != nil {
		return false
	}
	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
