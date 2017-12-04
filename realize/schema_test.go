package realize

import (
	"flag"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
	"testing"
)

func TestSchema_Add(t *testing.T) {
	r := Realize{}
	p := Project{Name: "test"}
	r.Add(p)
	if len(r.Schema.Projects) != 1 {
		t.Error("Unexpected error there are", len(r.Schema.Projects), "instead one")
	}
	r.Add(p)
	if len(r.Schema.Projects) != 1 {
		t.Error("Unexpected error there are", len(r.Schema.Projects), "instead one")
	}
	r.Add(Project{Name: "testing"})
	if len(r.Schema.Projects) != 2 {
		t.Error("Unexpected error there are", len(r.Schema.Projects), "instead two")
	}
}

func TestSchema_Remove(t *testing.T) {
	r := Realize{}
	r.Schema.Projects = []Project{
		{
			Name: "test",
		}, {
			Name: "testing",
		}, {
			Name: "testing",
		},
	}
	r.Remove("testing")
	if len(r.Schema.Projects) != 2 {
		t.Error("Unexpected errore there are", len(r.Schema.Projects), "instead one")
	}
}

func TestSchema_New(t *testing.T) {
	r := Realize{}
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
	set.Parse([]string{"--fmt", "--install", "--run", "--build", "--generate", "--test", "--vet"})
	p := r.New(c)
	if p.Name != filepath.Base(Wdir()) {
		t.Error("Unexpected error", p.Name, "instead", filepath.Base(Wdir()))
	}
	if !p.Tools.Install.Status {
		t.Error("Install should be enabled")
	}
	if !p.Tools.Fmt.Status {
		t.Error("Fmt should be enabled")
	}
	if !p.Tools.Run.Status {
		t.Error("Run should be enabled")
	}
	if !p.Tools.Build.Status {
		t.Error("Build should be enabled")
	}
	if !p.Tools.Generate.Status {
		t.Error("Generate should be enabled")
	}
	if !p.Tools.Test.Status {
		t.Error("Test should be enabled")
	}
	if !p.Tools.Vet.Status {
		t.Error("Vet should be enabled")
	}
}

func TestSchema_Filter(t *testing.T) {
	r := Realize{}
	r.Schema.Projects = []Project{
		{
			Name: "test",
		}, {
			Name: "test",
		},
		{
			Name: "example",
		},
	}
	result := r.Filter("Name", "test")
	if len(result) != 2 {
		t.Error("Expected two project")
	}
	result = r.Filter("Name", "example")
	if len(result) != 1 {
		t.Error("Expected one project")
	}
}
