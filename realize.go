package main

import (
	"errors"
	"fmt"
	s "github.com/tockins/realize/server"
	c "github.com/tockins/realize/settings"
	w "github.com/tockins/realize/watcher"
	"gopkg.in/urfave/cli.v2"
	"os"
)

const (
	name        = "Realize"
	version     = "1.3"
	description = "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths"
	config      = "realize.yaml"
	outputs     = "outputs.log"
	errs        = "errors.log"
	logs        = "logs.log"
	host        = "localhost"
	port        = 5001
)

var r realize

// Realize struct contains the general app informations
type realize struct {
	c.Settings                                      `yaml:"settings,omitempty"`
	Name, Description, Author, Email, Host, Version string       `yaml:"-"`
	Sync                                            chan string  `yaml:"-"`
	Blueprint                                       w.Blueprint  `yaml:"-"`
	Server                                          s.Server     `yaml:"-"`
	Projects                                        *[]w.Project `yaml:"projects" json:"projects"`
}

// Realize struct initialization
func init() {
	r = realize{
		Name:        name,
		Version:     version,
		Description: description,
		Sync:        make(chan string),
		Settings: c.Settings{
			Resources: c.Resources{
				Config:  config,
				Outputs: outputs,
				Logs:    logs,
				Errors:  errs,
			},
		},
	}
	r.Blueprint = w.Blueprint{
		Settings: &r.Settings,
		Sync:     r.Sync,
	}
	r.Server = s.Server{
		Blueprint: &r.Blueprint,
		Settings:  &r.Settings,
		Sync:      r.Sync,
	}
	r.Projects = &r.Blueprint.Projects

	// read if exist
	r.Read(&r)

	// increase the file limit
	if r.Config.Flimit != 0 {
		r.Flimit()
	}
}

// Before of every exec of a cli method
func before() error {
	fmt.Println(r.Blue.Bold(name) + " - " + r.Blue.Bold(version))
	fmt.Println(r.Blue.Regular(description) + "\n")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return handle(errors.New("$GOPATH isn't set up properly"))
	}
	return nil
}

// Handle errors
func handle(err error) error {
	if err != nil {
		fmt.Println(r.Red.Bold(err.Error()))
		os.Exit(1)
	}
	return nil
}

// Cli commands
func main() {
	c := &cli.App{
		Name:    r.Name,
		Version: r.Version,
		Authors: []*cli.Author{
			{
				Name:  "Alessio Pracchia",
				Email: "pracchia@hastegit",
			},
			{
				Name:  "Daniele Conventi",
				Email: "conventi@hastegit",
			},
		},
		Usage: r.Description,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run a toolchain on a project. Can be personalized, used with a single project and without make a realize config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path"},
					&cli.IntFlag{Name: "flimit", Aliases: []string{"f"}, Usage: "Increase files limit"},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Enable legacy watch"},
					&cli.IntFlag{Name: "legacy-delay", Aliases: []string{"ld"}, Usage: "Restarting delay for legacy watch"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "preview", Aliases: []string{"prev"}, Value: false, Usage: "Print each watched file"},
					&cli.BoolFlag{Name: "no-run", Aliases: []string{"nr"}, Usage: "Disable go run"},
					&cli.BoolFlag{Name: "no-bin", Aliases: []string{"nb"}, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-fmt", Aliases: []string{"nf"}, Usage: "Disable go fmt"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Run ignoring an existing config file"},
					&cli.BoolFlag{Name: "no-server", Aliases: []string{"ns"}, Value: false, Usage: "Disable web panel"},
					&cli.BoolFlag{Name: "serv-open", Aliases: []string{"so"}, Value: false, Usage: "Open wen panel in a new browser tab"},
					&cli.IntFlag{Name: "serv-port", Aliases: []string{"sp"}, Value: port, Usage: "Server port number"},
					&cli.StringFlag{Name: "serv-host", Aliases: []string{"sh"}, Value: host, Usage: "Server host"},
				},
				Action: func(p *cli.Context) error {
					r.Settings.Init(p)
					if r.Settings.Config.Create || len(r.Blueprint.Projects) <= 0 {
						r.Blueprint.Projects = []w.Project{}
						handle(r.Blueprint.Add(p))
					}
					handle(r.Server.Start(p))
					handle(r.Blueprint.Run())
					handle(r.Record(r))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "config",
				Category: "Configuration",
				Aliases:  []string{"c"},
				Usage:    "Create/Edit a realize config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: r.Wdir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "preview", Aliases: []string{"prev"}, Value: false, Usage: "Print each watched file"},
					&cli.BoolFlag{Name: "no-run", Aliases: []string{"nr"}, Usage: "Disable go run"},
					&cli.BoolFlag{Name: "no-bin", Aliases: []string{"nb"}, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-fmt", Aliases: []string{"nf"}, Usage: "Disable go fmt"},
				},
				Action: func(p *cli.Context) (err error) {
					handle(r.Blueprint.Insert(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully added."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "add",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Add a new project to an existing realize config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: r.Wdir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "preview", Aliases: []string{"prev"}, Value: false, Usage: "Print each watched file"},
					&cli.BoolFlag{Name: "no-run", Aliases: []string{"nr"}, Usage: "Disable go run"},
					&cli.BoolFlag{Name: "no-bin", Aliases: []string{"nb"}, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-fmt", Aliases: []string{"nf"}, Usage: "Disable go fmt"},
				},
				Action: func(p *cli.Context) (err error) {
					handle(r.Blueprint.Insert(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully added."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "remove",
				Category: "Configuration",
				Aliases:  []string{"r"},
				Usage:    "Remove a project from a config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
				},
				Action: func(p *cli.Context) error {
					handle(r.Blueprint.Remove(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully removed."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "list",
				Category: "Configuration",
				Aliases:  []string{"l"},
				Usage:    "Projects list",
				Action: func(p *cli.Context) error {
					return handle(r.Blueprint.List())
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "clean",
				Category: "Configuration",
				Aliases:  []string{"c"},
				Usage:    "Remove realize folder",
				Action: func(p *cli.Context) error {
					handle(r.Settings.Remove())
					fmt.Println(r.Green.Bold("Realize folder successfully removed."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
		},
	}
	c.Run(os.Args)
}
