package watcher

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"testing"
)

func TestArgsParam(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	params := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	result := argsParam(params)
	if len(result) != 2 {
		t.Fatal("Expected 2 instead", len(result))
	}
}

func TestDuplicates(t *testing.T) {
	projects := []Project{
		{
			Name: "a",
		}, {
			Name: "b",
		}, {
			Name: "c",
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
