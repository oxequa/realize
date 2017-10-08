package main

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestBlueprint_Clean(t *testing.T) {
	blp := Blueprint{}
	blp.Settings = &Settings{}
	blp.Projects = append(blp.Projects, Project{Name: "test0"})
	blp.Projects = append(blp.Projects, Project{Name: "test0"})
	blp.clean()
	if len(blp.Projects) > 1 {
		t.Error("Expected only one project")
	}
	blp.Projects = append(blp.Projects, Project{Path: "test1"})
	blp.Projects = append(blp.Projects, Project{Path: "test1"})
	blp.clean()
	if len(blp.Projects) > 2 {
		t.Error("Expected only one project")
	}

}

func TestBlueprint_Add(t *testing.T) {
	blp := Blueprint{}
	blp.Settings = &Settings{}
	// add all flags, test with expected
	set := flag.NewFlagSet("test", 0)
	set.Bool("fmt", false, "")
	set.Bool("vet", false, "")
	set.Bool("test", false, "")
	set.Bool("install", false, "")
	set.Bool("run", false, "")
	set.Bool("build", false, "")
	set.Bool("generate", false, "")
	set.String("path", "", "")
	c := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--path=test_path", "--fmt", "--install", "--run", "--build", "--generate", "--test", "--vet"})
	blp.add(c)
	expected := Project{
		Name: "test_path",
		Path: "test_path",
		Cmds: Cmds{
			Fmt: Cmd{
				Status: true,
			},
			Install: Cmd{
				Status: true,
			},
			Generate: Cmd{
				Status: true,
			},
			Test: Cmd{
				Status: true,
			},
			Build: Cmd{
				Status: true,
			},
			Vet: Cmd{
				Status: true,
			},
			Run: true,
		},
		Watcher: Watch{
			Paths:  []string{"/"},
			Ignore: []string{"vendor"},
			Exts:   []string{"go"},
		},
	}
	if !reflect.DeepEqual(blp.Projects[0], expected) {
		t.Error("Expected equal struct")
	}
}

func TestBlueprint_Check(t *testing.T) {
	blp := Blueprint{}
	blp.Settings = &Settings{}
	err := blp.check()
	if err == nil {
		t.Error("There is no project, error expected")
	}
	blp.Projects = append(blp.Projects, Project{Name: "test0"})
	err = blp.check()
	if err != nil {
		t.Error("There is a project, error unexpected", err)
	}
}

func TestBlueprint_Remove(t *testing.T) {
	blp := Blueprint{}
	blp.Settings = &Settings{}
	set := flag.NewFlagSet("name", 0)
	set.String("name", "", "")
	c := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--name=test0"})
	err := blp.remove(c)
	if err == nil {
		t.Error("Expected an error, there are no projects")
	}
	// Append a new project
	blp.Projects = append(blp.Projects, Project{Name: "test0"})
	err = blp.remove(c)
	if err != nil {
		t.Error("Error unexpected, the project should be remove", err)
	}
}

func TestBlueprint_Run(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	params := cli.NewContext(nil, set, nil)
	m := make(map[string]string)
	m["test"] = "test"
	projects := Blueprint{}
	projects.Settings = &Settings{}
	projects.Projects = []Project{
		{
			Name:        "test0",
			Path:        ".",
			Environment: m,
		},
		{
			Name: "test1",
			Path: ".",
		},
		{
			Name: "test1",
			Path: ".",
		},
	}
	go projects.run(params)
	if os.Getenv("test") != "test" {
		t.Error("Env variable seems different from that given", os.Getenv("test"), "expected", m["test"])
	}
	time.Sleep(5 * time.Second)
}
