package main

import (
	"errors"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
)

// Tool options customizable, should be moved in Cmd
type tool struct {
	dir, status     bool
	name, err    	string
	cmd, options    []string
}

// Cmds list of go commands
type Cmds struct {
	Fix      Cmd  `yaml:"fix,omitempty" json:"fix,omitempty"`
	Clean    Cmd  `yaml:"clean,omitempty" json:"clean,omitempty"`
	Vet      Cmd  `yaml:"vet,omitempty" json:"vet,omitempty"`
	Fmt      Cmd  `yaml:"fmt,omitempty" json:"fmt,omitempty"`
	Test     Cmd  `yaml:"test,omitempty" json:"test,omitempty"`
	Generate Cmd  `yaml:"generate,omitempty" json:"generate,omitempty"`
	Install  Cmd  `yaml:"install,omitempty" json:"install,omitempty"`
	Build    Cmd  `yaml:"build,omitempty" json:"build,omitempty"`
	Run      bool `yaml:"run,omitempty" json:"run,omitempty"`
}

// Cmd single command fields and options
type Cmd struct {
	Status                 bool     `yaml:"status,omitempty" json:"status,omitempty"`
	Method                 string   `yaml:"method,omitempty" json:"method,omitempty"`
	Args                   []string `yaml:"args,omitempty" json:"args,omitempty"`
	method                 []string
	tool                   bool
	name, startTxt, endTxt string
}

// Clean duplicate projects
func (r *realize) clean() error {
	if len(r.Schema) > 0 {
		arr := r.Schema
		for key, val := range arr {
			if _, err := duplicates(val, arr[key+1:]); err != nil {
				r.Schema = append(arr[:key], arr[key+1:]...)
				break
			}
		}
		return nil
	}
	return errors.New("there are no projects")
}

// Add a new project
func (r *realize) add(p *cli.Context)  (err error) {
	var path string
	// #118 get relative and if not exist try to get abs
	if _, err = os.Stat(p.String("path")); os.IsNotExist(err) {
		// path doesn't exist
		path, err = filepath.Abs(p.String("path"))
		if err != nil {
			return err
		}
	}else{
		path = filepath.Clean(p.String("path"))
	}

	project := Project{
		Name: filepath.Base(wdir()),
		Path: path,
		Cmds: Cmds{
			Vet: Cmd{
				Status: p.Bool("vet"),
			},
			Fmt: Cmd{
				Status: p.Bool("fmt"),
			},
			Test: Cmd{
				Status: p.Bool("test"),
			},
			Generate: Cmd{
				Status: p.Bool("generate"),
			},
			Build: Cmd{
				Status: p.Bool("build"),
			},
			Install: Cmd{
				Status: p.Bool("install"),
			},
			Run: p.Bool("run"),
		},
		Args: params(p),
		Watcher: Watch{
			Paths:  []string{"/"},
			Ignore: []string{".git", ".realize", "vendor"},
			Exts:   []string{"go"},
		},
	}
	if _, err := duplicates(project, r.Schema); err != nil {
		return err
	}
	r.Schema = append(r.Schema, project)
	return nil
}

// Run launches the toolchain for each project
func (r *realize) run(p *cli.Context) error {
	var match bool
	// check projects and remove duplicates
	if err := r.clean(); err != nil {
		return err
	}
	// set gobin
	if err := os.Setenv("GOBIN", filepath.Join(os.Getenv("GOPATH"), "bin")); err != nil {
		return err
	}
	// loop projects
	if p.String("name") != "" {
		wg.Add(1)
	} else {
		wg.Add(len(r.Schema))
	}
	for k, elm := range r.Schema {
		// command start using name flag
		if p.String("name") != "" && elm.Name != p.String("name") {
			continue
		}
		match = true
		r.Schema[k].config(r)
		go r.Schema[k].watch()
	}
	if !match {
		return errors.New("there is no project with the given name")
	}
	wg.Wait()
	return nil
}

// Remove a project
func (r *realize) remove(p *cli.Context) error {
	for key, val := range r.Schema {
		if p.String("name") == val.Name {
			r.Schema = append(r.Schema[:key], r.Schema[key+1:]...)
			return nil
		}
	}
	return errors.New("no project found")
}

// Insert current project if there isn't already one
func (r *realize) insert(c *cli.Context) error {
	if c.Bool("no-config") {
		r.Schema = []Project{}
	}
	if len(r.Schema) <= 0 {
		if err := r.add(c); err != nil {
			return err
		}
	}
	return nil
}
