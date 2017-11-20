package realize

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
	"testing"
)

func TestArgsParam(t *testing.T) {
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

func TestInArray(t *testing.T) {
	arr := []string{"a", "b", "c"}
	if !array(arr[0], arr) {
		t.Fatal("Unexpected", arr[0], "should be in", arr)
	}
	if array("d", arr) {
		t.Fatal("Unexpected", "d", "shouldn't be in", arr)
	}
}

func TestWdir(t *testing.T) {
	expected, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	result := wdir()
	if result != expected {
		t.Error("Expected", filepath.Base(expected), "instead", result)
	}
}
