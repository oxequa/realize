package server

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"runtime"
)

var cmd map[string]string
var stderr bytes.Buffer

// Init an associative array with the os supported
func init() {
	cmd = map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
}

// Open a url in the default browser
func Open(url string) (io.Writer, error) {
	open, err := cmd[runtime.GOOS];
	if !err {
		return nil, errors.New("This operating system is not supported.")
	}
	cmd := exec.Command(open, url)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return cmd.Stderr, err
	}

	return nil, nil
}
