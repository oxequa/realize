package main

import (
	"errors"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
	"time"
)

// Tool options customizable
type tool struct {
	dir     bool
	status  bool
	cmd     string
	name    string
	err     string
	options []string
}

// Cmds
type Cmds struct {
	Vet      Cmd  `yaml:"vet,omitempty" json:"vet,omitempty"`
	Fmt      Cmd  `yaml:"fmt,omitempty" json:"fmt,omitempty"`
	Test     Cmd  `yaml:"test,omitempty" json:"test,omitempty"`
	Generate Cmd  `yaml:"generate,omitempty" json:"generate,omitempty"`
	Install  Cmd  `yaml:"install" json:"install"`
	Build    Cmd  `yaml:"build,omitempty" json:"build,omitempty"`
	Run      bool `yaml:"run,omitempty" json:"run,omitempty"`
}

// Cmd
type Cmd struct {
	Status                 bool     `yaml:"status,omitempty" json:"status,omitempty"`
	Args                   []string `yaml:"args,omitempty" json:"args,omitempty"`
	name, startTxt, endTxt string
}

// Clean duplicate projects
func (r *realize) clean() {
	arr := r.Schema
	for key, val := range arr {
		if _, err := duplicates(val, arr[key+1:]); err != nil {
			r.Schema = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Check whether there is a project
func (r *realize) check() error {
	if len(r.Schema) > 0 {
		r.clean()
		return nil
	}
	return errors.New("there are no projects")
}

// Add a new project
func (r *realize) add(p *cli.Context) error {
	project := Project{
		Name: r.Settings.name(p.String("name"), p.String("path")),
		Path: r.Settings.path(p.String("path")),
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
			Ignore: []string{"vendor"},
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
	err := r.check()
	if err == nil {
		// loop projects
		if p.String("name") != "" {
			wg.Add(1)
		} else {
			wg.Add(len(r.Schema))
		}
		for k, elm := range r.Schema {
			if p.String("name") != "" && r.Schema[k].Name != p.String("name") {
				continue
			}
			if elm.Cmds.Fmt.Status {
				if len(elm.Cmds.Fmt.Args) == 0 {
					elm.Cmds.Fmt.Args = []string{"-s", "-w", "-e", "./.."}
				}
				r.Schema[k].tools = append(r.Schema[k].tools, tool{
					status:  elm.Cmds.Fmt.Status,
					cmd:     "gofmt",
					options: split([]string{}, elm.Cmds.Fmt.Args),
					name:    "Go Fmt",
				})
			}
			if elm.Cmds.Generate.Status {
				r.Schema[k].tools = append(r.Schema[k].tools, tool{
					status:  elm.Cmds.Generate.Status,
					cmd:     "go",
					options: split([]string{"generate"}, elm.Cmds.Generate.Args),
					name:    "Go Generate",
					dir:     true,
				})
			}
			if elm.Cmds.Test.Status {
				r.Schema[k].tools = append(r.Schema[k].tools, tool{
					status:  elm.Cmds.Test.Status,
					cmd:     "go",
					options: split([]string{"test"}, elm.Cmds.Test.Args),
					name:    "Go Test",
					dir:     true,
				})
			}
			if elm.Cmds.Vet.Status {
				r.Schema[k].tools = append(r.Schema[k].tools, tool{
					status:  elm.Cmds.Vet.Status,
					cmd:     "go",
					options: split([]string{"vet"}, elm.Cmds.Vet.Args),
					name:    "Go Vet",
					dir:     true,
				})
			}
			// default settings
			r.Schema[k].Cmds.Install = Cmd{
				Status:   elm.Cmds.Install.Status,
				Args:     append([]string{"install"}, elm.Cmds.Install.Args...),
				name:     "Go Install",
				startTxt: "Instaling...",
				endTxt:   "Installed",
			}
			r.Schema[k].Cmds.Build = Cmd{
				Status:   elm.Cmds.Build.Status,
				Args:     append([]string{"build"}, elm.Cmds.Build.Args...),
				name:     "Go Build",
				startTxt: "Bulding...",
				endTxt:   "Built",
			}

			r.Schema[k].parent = r
			r.Schema[k].path = r.Schema[k].Path

			// env variables
			for key, item := range r.Schema[k].Environment {
				if err := os.Setenv(key, item); err != nil {
					r.Schema[k].Buffer.StdErr = append(r.Schema[k].Buffer.StdErr, BufferOut{Time: time.Now(), Text: err.Error(), Type: "Env error", Stream: ""})
				}
			}

			// base path of the project
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if elm.path == "." || elm.path == "/" {
				r.Schema[k].base = wd
				r.Schema[k].path = elm.wdir()
			} else if filepath.IsAbs(elm.path) {
				r.Schema[k].base = elm.path
			} else {
				r.Schema[k].base = filepath.Join(wd, elm.path)
			}
			go r.Schema[k].watch()
		}
		wg.Wait()
		return nil
	}
	return err
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

// Insert a project if there isn't already one
func (r *realize) insert(c *cli.Context) error {
	if !c.Bool("config") {
		r.Schema = []Project{}
	}
	if len(r.Schema) <= 0 {
		if err := r.add(c); err != nil {
			return err
		}
	}
	return nil
}
