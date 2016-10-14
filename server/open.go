package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

var cli map[string]string
var stderr bytes.Buffer

func init() {
	cli = map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
}

func Open(url string) (io.Writer, error) {
	if open, err := cli[runtime.GOOS]; !err {
		return nil, errors.New("This operating system is not supported.")
	} else {
		cmd := exec.Command(open, url)
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(cmd.Stderr, err)
			return cmd.Stderr, err
		}
	}
	return nil, nil
}
