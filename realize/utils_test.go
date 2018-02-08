package realize

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
	"testing"
)

func TestParams(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	p := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	result := params(p)
	if len(result) != 2 {
		t.Fatal("Expected 2 instead", len(result))
	}
}

func TestDuplicates(t *testing.T) {
	projects := []Project{
		{
			Name: "a",
			Path: "a",
		}, {
			Name: "a",
			Path: "a",
		}, {
			Name: "c",
			Path: "c",
		},
	}
	_, err := duplicates(projects[0], projects)
	if err == nil {
		t.Fatal("Error unexpected", err)
	}
	_, err = duplicates(Project{}, projects)
	if err != nil {
		t.Fatal("Error unexpected", err)
	}

}

func TestWdir(t *testing.T) {
	expected, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	result := Wdir()
	if result != expected {
		t.Error("Expected", filepath.Base(expected), "instead", result)
	}
}

func TestExt(t *testing.T) {
	paths := map[string]string{
		"/test/a/b/c":        "",
		"/test/a/ac.go":      "go",
		"/test/a/ac.test.go": "go",
		"/test/a/ac_test.go": "go",
		"/test/./ac_test.go": "go",
		"/test/a/.test":      "test",
		"/test/a/.":          "",
	}
	for i, v := range paths {
		if ext(i) != v {
			t.Error("Wrong extension", ext(i), v)
		}
	}

}
