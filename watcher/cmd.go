package cli

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
	"strings"
)

// Watch method adds the given paths on the Watcher
func (h *Blueprint) Run() error {
	err := h.check()
	if err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			h.Projects[k].parent = h
			go h.Projects[k].watching()
		}
		wg.Wait()
		return nil
	}
	return err
}

// Fast method run a project from his working directory without makes a config file
func (h *Blueprint) Fast(p *cli.Context) error {
	// Takes the values from config if wd path match with someone else
	wg.Add(1)
	for i := 0; i < len(h.Projects); i++ {
		v := &h.Projects[i]
		v.parent = h
		v.path = v.Path
		go v.watching()
	}
	wg.Wait()
	return nil
}

// Add a new project
func (h *Blueprint) Add(p *cli.Context) error {
	project := Project{
		Name:   h.name(p),
		Path:   filepath.Clean(p.String("path")),
		Build:  p.Bool("build"),
		Bin:    boolFlag(p.Bool("no-bin")),
		Run:    boolFlag(p.Bool("no-run")),
		Fmt:    boolFlag(p.Bool("no-fmt")),
		Test:   p.Bool("test"),
		Params: argsParam(p),
		Watcher: Watcher{
			Paths:  []string{"/"},
			Ignore: []string{"vendor"},
			Exts:   []string{".go"},
			Output: map[string]bool{
				"cli":  true,
				"file": false,
			},
		},
	}
	if _, err := duplicates(project, h.Projects); err != nil {
		return err
	}
	h.Projects = append(h.Projects, project)
	return nil
}

// Clean duplicate projects
func (h *Blueprint) Clean() {
	arr := h.Projects
	for key, val := range arr {
		if _, err := duplicates(val, arr[key+1:]); err != nil {
			h.Projects = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Inserts a new project in the list
func (h *Blueprint) Insert(p *cli.Context) error {
	err := h.Add(p)
	return err
}

// Remove a project
func (h *Blueprint) Remove(p *cli.Context) error {
	for key, val := range h.Projects {
		if p.String("name") == val.Name {
			h.Projects = append(h.Projects[:key], h.Projects[key+1:]...)
			return nil
		}
	}
	return errors.New("No project found.")
}

// List of all the projects
func (h *Blueprint) List() error {
	err := h.check()
	if err == nil {
		for _, val := range h.Projects {
			fmt.Println(h.Blue.Bold("|"), h.Blue.Bold(strings.ToUpper(val.Name)))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Base Path"), ":", h.Magenta.Regular(val.Path))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Run"), ":", h.Magenta.Regular(val.Run))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Build"), ":", h.Magenta.Regular(val.Build))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Install"), ":", h.Magenta.Regular(val.Bin))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Fmt"), ":", h.Magenta.Regular(val.Fmt))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Test"), ":", h.Magenta.Regular(val.Test))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Params"), ":", h.Magenta.Regular(val.Params))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Watcher"), ":")
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("After"), ":", h.Magenta.Regular(val.Watcher.After))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Before"), ":", h.Magenta.Regular(val.Watcher.Before))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Extensions"), ":", h.Magenta.Regular(val.Watcher.Exts))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Paths"), ":", h.Magenta.Regular(val.Watcher.Paths))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Paths ignored"), ":", h.Magenta.Regular(val.Watcher.Ignore))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Watch preview"), ":", h.Magenta.Regular(val.Watcher.Preview))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Output"), ":")
			fmt.Println(h.Magenta.Regular("|"), "\t\t\t", h.Yellow.Regular("Cli"), ":", h.Magenta.Regular(val.Watcher.Output["cli"]))
			fmt.Println(h.Magenta.Regular("|"), "\t\t\t", h.Yellow.Regular("File"), ":", h.Magenta.Regular(val.Watcher.Output["file"]))
		}
		return nil
	}
	return err
}

// Check if there are projects
func (h *Blueprint) check() error {
	if len(h.Projects) > 0 {
		h.Clean()
		return nil
	} else {
		return errors.New("There are no projects. The config file is empty.")
	}
}

// NameParam check the project name presence. If empty takes the working directory name
func (h *Blueprint) name(p *cli.Context) string {
	var name string
	if p.String("name") == "" && p.String("path") == "" {
		return h.Wdir()
	} else if p.String("path") != "/" {
		name = filepath.Base(p.String("path"))
	} else {
		name = p.String("name")
	}
	return name
}
