package main

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"log"
	"reflect"
	"testing"
	"time"
)

type logT struct{}

func (logT) Write(bytes []byte) (int, error) {
	return 0, nil
}

func TestMain(m *testing.M) {
	log.SetFlags(0)
	log.SetOutput(logT{})
}

func TestBlueprint_Clean(t *testing.T) {
	r := realize{}
	r.Schema = append(r.Schema, Project{Name: "test0"})
	r.Schema = append(r.Schema, Project{Name: "test0"})
	r.clean()
	if len(r.Schema) > 1 {
		t.Error("Expected only one project")
	}
	r.Schema = append(r.Schema, Project{Path: "test1"})
	r.Schema = append(r.Schema, Project{Path: "test1"})
	r.clean()
	if len(r.Schema) > 2 {
		t.Error("Expected only one project")
	}

}

func TestBlueprint_Add(t *testing.T) {
	r := realize{}
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
	r.add(c)
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
	if !reflect.DeepEqual(r.Schema[0], expected) {
		t.Error("Expected equal struct")
	}
}

func TestBlueprint_Check(t *testing.T) {
	r := realize{}
	err := r.check()
	if err == nil {
		t.Error("There is no project, error expected")
	}
	r.Schema = append(r.Schema, Project{Name: "test0"})
	err = r.check()
	if err != nil {
		t.Error("There is a project, error unexpected", err)
	}
}

func TestBlueprint_Remove(t *testing.T) {
	r := realize{}
	set := flag.NewFlagSet("name", 0)
	set.String("name", "", "")
	c := cli.NewContext(nil, set, nil)
	set.Parse([]string{"--name=test0"})
	err := r.remove(c)
	if err == nil {
		t.Error("Expected an error, there are no projects")
	}
	// Append a new project
	r.Schema = append(r.Schema, Project{Name: "test0"})
	err = r.remove(c)
	if err != nil {
		t.Error("Error unexpected, the project should be remove", err)
	}
}

func TestBlueprint_Run(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	params := cli.NewContext(nil, set, nil)
	m := make(map[string]string)
	m["test"] = "test"
	r := realize{}
	r.Schema = []Project{
		{
			Name: "test0",
			Path: ".",
		},
		{
			Name: "test1",
			Path: "/test",
		},
		{
			Name: "test2",
			Path: "/test",
		},
	}
	go r.run(params)
	time.Sleep(1 * time.Second)
}
