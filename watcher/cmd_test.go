package watcher

import (
	"flag"
	"github.com/tockins/realize/settings"
	cli "gopkg.in/urfave/cli.v2"
	"testing"
	"time"
)

func TestBlueprint_Run(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	params := cli.NewContext(nil, set, nil)
	projects := Blueprint{}
	projects.Settings = &settings.Settings{}
	projects.Projects = []Project{
		{
			Name: "test1",
			Path: ".",
		},
		{
			Name: "test1",
			Path: ".",
		},
		{
			Name: "test2",
			Path: ".",
		},
	}
	go projects.Run(params)
	time.Sleep(100 * time.Millisecond)
}

func TestBlueprint_Add(t *testing.T) {
	projects := Blueprint{}
	projects.Settings = &settings.Settings{}
	// add all flags, test with expected
	set := flag.NewFlagSet("test", 0)
	set.String("name", "default_name", "doc")
	set.String("path", "default_path", "doc")
	params := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--name", "name", "name"})
	set.Parse([]string{"--path", "path", "path"})
	projects.Add(params)
}
