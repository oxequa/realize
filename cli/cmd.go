package cli

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strings"
)

// Watch method adds the given paths on the Watcher
func (h *Blueprint) Run() error {
	err := h.Read()
	if err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			go h.Projects[k].watching()
		}
		wg.Wait()
		return nil
	}
	return err
}

// Fast method run a project from his working directory without makes a config file
func (h *Blueprint) Fast(params *cli.Context) error {
	fast := h.Projects[0]
	// Takes the values from config if wd path match with someone else
	if params.Bool("config") {
		if err := h.Read(); err == nil {
			for _, val := range h.Projects {
				if fast.Path == val.Path {
					fast = val
				}
			}
		}
	}
	wg.Add(1)
	go fast.watching()
	wg.Wait()
	return nil
}

// Add a new project
func (h *Blueprint) Add(params *cli.Context) error {
	p := Project{
		Name:   nameFlag(params),
		Path:   filepath.Clean(params.String("path")),
		Build:  params.Bool("build"),
		Bin:    boolFlag(params.Bool("no-bin")),
		Run:    boolFlag(params.Bool("no-run")),
		Fmt:    boolFlag(params.Bool("no-fmt")),
		Test:   params.Bool("test"),
		Params: argsParam(params),
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
	if _, err := duplicates(p, h.Projects); err != nil {
		return err
	}
	h.Projects = append(h.Projects, p)
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

// Read, Check and remove duplicates from the config file
func (h *Blueprint) Read() error {
	content, err := read(h.Files["config"])
	if err == nil {
		err = yaml.Unmarshal(content, h)
		if err == nil {
			if len(h.Projects) > 0 {
				h.Clean()
				return nil
			}
			return errors.New("There are no projects!")
		}
		return err
	}
	return err
}

// Create and unmarshal yaml config file
func (h *Blueprint) Create() error {
	y, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	return write(h.Files["config"], y)
}

// Inserts a new project in the list
func (h *Blueprint) Insert(params *cli.Context) error {
	check := h.Read()
	err := h.Add(params)
	if err == nil {
		err = h.Create()
		if check == nil && err == nil {
			fmt.Println(Green("Your project was successfully added"))
		} else {
			fmt.Println(Green("The config file was successfully created"))
		}
	}
	return err
}

// Remove a project
func (h *Blueprint) Remove(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		for key, val := range h.Projects {
			if params.String("name") == val.Name {
				h.Projects = append(h.Projects[:key], h.Projects[key+1:]...)
				err = h.Create()
				if err == nil {
					fmt.Println(Green("Your project was successfully removed"))
				}
				return err
			}
		}
		return errors.New("No project found")
	}
	return err
}

// List of all the projects
func (h *Blueprint) List() error {
	err := h.Read()
	if err == nil {
		for _, val := range h.Projects {
			fmt.Println(Blue("|"), Blue(strings.ToUpper(val.Name)))
			fmt.Println(MagentaS("|"), "\t", Yellow("Base Path"), ":", MagentaS(val.Path))
			fmt.Println(MagentaS("|"), "\t", Yellow("Run"), ":", MagentaS(val.Run))
			fmt.Println(MagentaS("|"), "\t", Yellow("Build"), ":", MagentaS(val.Build))
			fmt.Println(MagentaS("|"), "\t", Yellow("Install"), ":", MagentaS(val.Bin))
			fmt.Println(MagentaS("|"), "\t", Yellow("Fmt"), ":", MagentaS(val.Fmt))
			fmt.Println(MagentaS("|"), "\t", Yellow("Test"), ":", MagentaS(val.Test))
			fmt.Println(MagentaS("|"), "\t", Yellow("Params"), ":", MagentaS(val.Params))
			fmt.Println(MagentaS("|"), "\t", Yellow("Watcher"), ":")
			fmt.Println(MagentaS("|"), "\t\t", Yellow("After"), ":", MagentaS(val.Watcher.After))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Before"), ":", MagentaS(val.Watcher.Before))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Extensions"), ":", MagentaS(val.Watcher.Exts))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths"), ":", MagentaS(val.Watcher.Paths))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths ignored"), ":", MagentaS(val.Watcher.Ignore))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Watch preview"), ":", MagentaS(val.Watcher.Preview))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Output"), ":")
			fmt.Println(MagentaS("|"), "\t\t\t", Yellow("Cli"), ":", MagentaS(val.Watcher.Output["cli"]))
			fmt.Println(MagentaS("|"), "\t\t\t", Yellow("File"), ":", MagentaS(val.Watcher.Output["file"]))
		}
	}
	return err
}
