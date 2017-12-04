package realize

import (
	"net/http"
	"runtime"
	"testing"
)

func TestServer_Start(t *testing.T) {
	s := Server{
		Host: "localhost",
		Port: 5000,
	}
	host := "http://localhost:5000/"
	urls := []string{
		host,
		host + "assets/js/all.min.js",
		host + "assets/css/app.css",
		host + "app/components/settings/index.html",
		host + "app/components/project/index.html",
		host + "app/components/project/index.html",
		host + "app/components/index.html",
	}
	err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	for _, elm := range urls {
		resp, err := http.Get(elm)
		if err != nil || resp.StatusCode != 200 {
			t.Fatal(err, resp.StatusCode, elm)
		}
	}
}

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
