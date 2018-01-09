package realize

import (
	"runtime"
	"testing"
)

func TestServer_Open(t *testing.T) {
	cmd := map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
	key := runtime.GOOS
	if _, ok := cmd[key]; !ok {
		t.Error("System not supported")
	}
}
