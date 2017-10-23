package main

import (
	"errors"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
	"time"
	"reflect"
	"fmt"
)

// Tool options customizable, should be moved in Cmd
type tool struct {
	dir     bool
	status  bool
	name    string
	err     string
	cmd     []string
	options []string
}

// Cmds list of go commands
type Cmds struct {
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
func (r *realize) clean() {
	arr := r.Schema
	for key, val := range arr {
		if _, err := duplicates(val, arr[key+1:]); err != nil {
			r.Schema = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Add a new project
func (r *realize) add(p *cli.Context) error {
	project := Project{
		Name: filepath.Base(filepath.Clean(p.String("path"))),
		Path: filepath.Clean(p.String("path")),
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
			Ignore: []string{".git",".realize","vendor"},
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
	if len(r.Schema) > 0 {
		r.clean()
	}else{
		return errors.New("there are no projects")
	}
	// set gobin
	err := os.Setenv("GOBIN", filepath.Join(os.Getenv("GOPATH"), "bin"))
	if err != nil {
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
		if p.String("name") != "" && r.Schema[k].Name != p.String("name") {
			continue
		}
		// validate project path, if invalid get wdir or clean current
		if !filepath.IsAbs(elm.Path){
			r.Schema[k].Path = wdir()
		}else{
			r.Schema[k].Path = filepath.Clean(elm.Path)
		}
		// env variables
		for key, item := range r.Schema[k].Environment {
			if err := os.Setenv(key, item); err != nil {
				r.Schema[k].Buffer.StdErr = append(r.Schema[k].Buffer.StdErr, BufferOut{Time: time.Now(), Text: err.Error(), Type: "Env error", Stream: ""})
			}
		}
		// get basepath name
		r.Schema[k].name = filepath.Base(r.Schema[k].Path)

		fields := reflect.Indirect(reflect.ValueOf(&r.Schema[k].Cmds))
		// Loop struct Cmds fields
		for i := 0; i < fields.NumField(); i++ {
			field := fields.Type().Field(i).Name
			if fields.FieldByName(field).Type().Name() == "Cmd" {
				v := fields.FieldByName(field)
				// Loop struct Cmd
				for i := 0; i < v.NumField(); i++ {
					//f := v.Type().Field(i).Name
					//fmt.Println(f)
					//if f.IsValid() {
					//	if f.CanSet() {
					//		fmt.Println(f.)
					//		//switch f.Kind() {
					//		//case reflect.Bool:
					//		//case reflect.String:
					//		//case reflect.Slice:
					//		//}
					//	}
					//}
				}
			}
		}

		if elm.Cmds.Fmt.Status {
			if len(elm.Cmds.Fmt.Args) == 0 {
				elm.Cmds.Fmt.Args = []string{"-s", "-w", "-e", "./"}
			}
			r.Schema[k].tools = append(r.Schema[k].tools, tool{
				status:  elm.Cmds.Fmt.Status,
				cmd:     replace([]string{"gofmt"}, r.Schema[k].Cmds.Fmt.Method),
				options: split([]string{}, elm.Cmds.Fmt.Args),
				name:    "Fmt",
			})
		}
		if elm.Cmds.Generate.Status {
			r.Schema[k].tools = append(r.Schema[k].tools, tool{
				status:  elm.Cmds.Generate.Status,
				cmd:     replace([]string{"go", "generate"}, r.Schema[k].Cmds.Generate.Method),
				options: split([]string{}, elm.Cmds.Generate.Args),
				name:    "Generate",
				dir:     true,
			})
		}
		if elm.Cmds.Test.Status {
			r.Schema[k].tools = append(r.Schema[k].tools, tool{
				status:  elm.Cmds.Test.Status,
				cmd:     replace([]string{"go", "test"}, r.Schema[k].Cmds.Test.Method),
				options: split([]string{}, elm.Cmds.Test.Args),
				name:    "Test",
				dir:     true,
			})
		}
		if elm.Cmds.Vet.Status {
			r.Schema[k].tools = append(r.Schema[k].tools, tool{
				status:  elm.Cmds.Vet.Status,
				cmd:     replace([]string{"go", "vet"}, r.Schema[k].Cmds.Vet.Method),
				options: split([]string{}, elm.Cmds.Vet.Args),
				name:    "Vet",
				dir:     true,
			})
		}
		// default settings
		r.Schema[k].Cmds.Install = Cmd{
			Status:   elm.Cmds.Install.Status,
			Args:     append([]string{}, elm.Cmds.Install.Args...),
			method:   replace([]string{"go", "install"}, r.Schema[k].Cmds.Install.Method),
			name:     "Install",
			startTxt: "Installing...",
			endTxt:   "Installed",
		}
		r.Schema[k].Cmds.Build = Cmd{
			Status:   elm.Cmds.Build.Status,
			Args:     append([]string{}, elm.Cmds.Build.Args...),
			method:   replace([]string{"go", "build"}, r.Schema[k].Cmds.Build.Method),
			name:     "Build",
			startTxt: "Building...",
			endTxt:   "Built",
		}
		r.Schema[k].parent = r

		match = true
		go r.Schema[k].watch()
	}
	if !match {
		return errors.New("there is no project with the given name")
	}
	wg.Wait()
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
