package watcher

import (
	"errors"
	"fmt"
	"github.com/tockins/realize/style"
	cli "gopkg.in/urfave/cli.v2"
	"strings"
)

// Run launches the toolchain for each project
func (h *Blueprint) Run() error {
	err := h.check()
	if err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k, element := range h.Projects {
			tools := tools{}
			if element.Cmds.Fmt {
				tools.Fmt = tool{
					status:  &h.Projects[k].Cmds.Fmt,
					cmd:     "gofmt",
					options: []string{"-s", "-w", "-e"},
					name:    "Go Fmt",
				}
			}
			if element.Cmds.Generate {
				tools.Generate = tool{
					status:  &h.Projects[k].Cmds.Generate,
					cmd:     "go",
					options: []string{"generate"},
					name:    "Go Generate",
				}
			}
			if element.Cmds.Test {
				tools.Test = tool{
					status:  &h.Projects[k].Cmds.Test,
					cmd:     "go",
					options: []string{"test"},
					name:    "Go Test",
				}
			}
			if element.Cmds.Vet {
				tools.Vet = tool{
					status:  &h.Projects[k].Cmds.Vet,
					cmd:     "go",
					options: []string{"test"},
					name:    "Go Test",
				}
			}
			h.Projects[k].tools = tools
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
		Name: h.Name(p.String("name"), p.String("path")),
		Path: h.Path(p.String("path")),
		Cmds: Cmds{

			Vet:      p.Bool("vet"),
			Fmt:      !p.Bool("no-fmt"),
			Test:     p.Bool("test"),
			Generate: p.Bool("generate"),
			Build: Cmd{
				Status: p.Bool("build"),
			},
			Bin: Cmd{
				Status: !p.Bool("no-bin"),
			},
			Run: !p.Bool("no-run"),
		},
		Args: argsParam(p),
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

// Remove a project
func (h *Blueprint) Remove(p *cli.Context) error {
	for key, val := range h.Projects {
		if p.String("name") == val.Name {
			h.Projects = append(h.Projects[:key], h.Projects[key+1:]...)
			return nil
		}
	}
	return errors.New("no project found")
}

// List of all the projects
func (h *Blueprint) List() error {
	err := h.check()
	if err == nil {
		for _, val := range h.Projects {
			fmt.Println(style.Blue.Bold("[") + strings.ToUpper(val.Name) + style.Blue.Bold("]"))
			name := style.Magenta.Bold("[") + strings.ToUpper(val.Name) + style.Magenta.Bold("]")

			fmt.Println(name, style.Yellow.Regular("Base Path"), ":", style.Magenta.Regular(val.Path))
			fmt.Println(name, style.Yellow.Regular("Fmt"), ":", style.Magenta.Regular(val.Cmds.Fmt))
			fmt.Println(name, style.Yellow.Regular("Generate"), ":", style.Magenta.Regular(val.Cmds.Generate))
			fmt.Println(name, style.Yellow.Regular("Test"), ":", style.Magenta.Regular(val.Cmds.Test))
			fmt.Println(name, style.Yellow.Regular("Install"), ":", style.Magenta.Regular(val.Cmds.Bin))
			fmt.Println(name, style.Yellow.Regular("Build"), ":", style.Magenta.Regular(val.Cmds.Build))
			fmt.Println(name, style.Yellow.Regular("Run"), ":", style.Magenta.Regular(val.Cmds.Run))
			if len(val.Args) > 0 {
				fmt.Println(name, style.Yellow.Regular("Params"), ":", style.Magenta.Regular(val.Args))
			}
			fmt.Println(name, style.Yellow.Regular("Watcher"), ":")
			fmt.Println(name, "\t", style.Yellow.Regular("Preview"), ":", style.Magenta.Regular(val.Watcher.Preview))
			if len(val.Watcher.Exts) > 0 {
				fmt.Println(name, "\t", style.Yellow.Regular("Extensions"), ":", style.Magenta.Regular(val.Watcher.Exts))
			}
			if len(val.Watcher.Paths) > 0 {
				fmt.Println(name, "\t", style.Yellow.Regular("Paths"), ":", style.Magenta.Regular(val.Watcher.Paths))
			}
			if len(val.Watcher.Ignore) > 0 {
				fmt.Println(name, "\t", style.Yellow.Regular("Ignored paths"), ":", style.Magenta.Regular(val.Watcher.Ignore))
			}
			if len(val.Watcher.Scripts) > 0 {
				fmt.Println(name, "\t", style.Yellow.Regular("Scripts"), ":")
				for _, v := range val.Watcher.Scripts {
					if v.Command != "" {
						fmt.Println(name, "\t\t", style.Magenta.Regular("-"), style.Yellow.Regular("Command"), ":", style.Magenta.Regular(v.Command))
						if v.Path != "" {
							fmt.Println(name, "\t\t", style.Yellow.Regular("Path"), ":", style.Magenta.Regular(v.Path))
						}
						if v.Type != "" {
							fmt.Println(name, "\t\t", style.Yellow.Regular("Type"), ":", style.Magenta.Regular(v.Type))
						}
					}
				}
			}
			fmt.Println(name, style.Yellow.Regular("Streams"), ":")
			fmt.Println(name, "\t", style.Yellow.Regular("Cli Out"), ":", style.Magenta.Regular(val.Streams.CliOut))
			fmt.Println(name, "\t", style.Yellow.Regular("File Out"), ":", style.Magenta.Regular(val.Streams.FileOut))
			fmt.Println(name, "\t", style.Yellow.Regular("File Log"), ":", style.Magenta.Regular(val.Streams.FileLog))
			fmt.Println(name, "\t", style.Yellow.Regular("File Err"), ":", style.Magenta.Regular(val.Streams.FileErr))
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
	return errors.New("there are no projects")
}
