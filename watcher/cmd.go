package cli

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
	"strings"
)

// Run launches the toolchain for each project
func (h *Blueprint) Run() error {
	err := h.check()
	if err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			h.Projects[k].parent = h
			h.Projects[k].path = h.Projects[k].Path
			if h.Legacy.Status {
				go h.Projects[k].watchByPolling()
			} else {
				go h.Projects[k].watchByNotify()
			}
		}
		wg.Wait()
		return nil
	}
	return err
}

// Add a new project
func (h *Blueprint) Add(p *cli.Context) error {
	project := Project{
		Name:     h.name(p),
		Path:     strings.Replace(filepath.Clean(p.String("path")), "\\", "/", -1),
		Fmt:      !p.Bool("no-fmt"),
		Generate: p.Bool("generate"),
		Test:     p.Bool("test"),
		Build:    p.Bool("build"),
		Bin:      !p.Bool("no-bin"),
		Run:      !p.Bool("no-run"),
		Params:   argsParam(p),
		Watcher: Watcher{
			Paths:   []string{"/"},
			Ignore:  []string{"vendor"},
			Exts:    []string{".go"},
			Preview: p.Bool("preview"),
			Scripts: []Command{},
		},
		Streams: Streams{
			CliOut:  true,
			FileOut: false,
			FileLog: false,
			FileErr: false,
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

// Insert a new project in projects list
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
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Fmt"), ":", h.Magenta.Regular(val.Fmt))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Generate"), ":", h.Magenta.Regular(val.Generate))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Test"), ":", h.Magenta.Regular(val.Test))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Install"), ":", h.Magenta.Regular(val.Bin))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Build"), ":", h.Magenta.Regular(val.Build))
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Run"), ":", h.Magenta.Regular(val.Run))
			if len(val.Params) > 0 {
				fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Params"), ":", h.Magenta.Regular(val.Params))
			}
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Watcher"), ":")
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Preview"), ":", h.Magenta.Regular(val.Watcher.Preview))
			if len(val.Watcher.Exts) > 0 {
				fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Extensions"), ":", h.Magenta.Regular(val.Watcher.Exts))
			}
			if len(val.Watcher.Paths) > 0 {
				fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Paths"), ":", h.Magenta.Regular(val.Watcher.Paths))
			}
			if len(val.Watcher.Ignore) > 0 {
				fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Ignored paths"), ":", h.Magenta.Regular(val.Watcher.Ignore))
			}
			if len(val.Watcher.Scripts) > 0 {
				fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Scripts"), ":")
				for _, v := range val.Watcher.Scripts {
					if v.Command != "" {
						fmt.Println(h.Magenta.Regular("|"), "\t\t\t", h.Magenta.Regular("-"), h.Yellow.Regular("Command"), ":", h.Magenta.Regular(v.Command))
						if v.Path != "" {
							fmt.Println(h.Magenta.Regular("|"), "\t\t\t", h.Yellow.Regular("Path"), ":", h.Magenta.Regular(v.Path))
						}
						if v.Type != "" {
							fmt.Println(h.Magenta.Regular("|"), "\t\t\t", h.Yellow.Regular("Type"), ":", h.Magenta.Regular(v.Type))
						}
					}
				}
			}
			fmt.Println(h.Magenta.Regular("|"), "\t", h.Yellow.Regular("Streams"), ":")
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("Cli Out"), ":", h.Magenta.Regular(val.Streams.CliOut))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("File Out"), ":", h.Magenta.Regular(val.Streams.FileOut))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("File Log"), ":", h.Magenta.Regular(val.Streams.FileLog))
			fmt.Println(h.Magenta.Regular("|"), "\t\t", h.Yellow.Regular("File Err"), ":", h.Magenta.Regular(val.Streams.FileErr))
		}
		return nil
	}
	return err
}

// Check whether there is a project
func (h *Blueprint) check() error {
	if len(h.Projects) > 0 {
		h.Clean()
		return nil
	}
	return errors.New("There are no projects. The config file is empty.")
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
