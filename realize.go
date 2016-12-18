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
	version     = "1.2.1"
	description = "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths"
	config      = "realize.yaml"
	streams     = "streams.log"
	errs        = "errors.log"
	logs        = "logs.log"
	host        = "localhost"
	port        = 5001
	server      = true
	open        = false
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
			Config: c.Config{
				Flimit: 0,
			},
			Resources: c.Resources{
				Config:  config,
				Streams: streams,
				Logs:    logs,
				Errors:  errs,
			},
			Server: c.Server{
				Enabled: server,
				Open:    open,
				Host:    host,
				Port:    port,
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
				Name:  "run",
				Usage: "Build and watch file changes. Can be used even with a single project or without the config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enables the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "no-server", Usage: "Disables the web panel"},
					&cli.BoolFlag{Name: "no-config", Value: false, Usage: "Uses the config settings"},
					&cli.BoolFlag{Name: "open", Usage: "Automatically opens the web panel"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enables the tests"},
				},
				Action: func(p *cli.Context) error {
					if p.Bool("no-config") {
						r.Settings = c.Settings{
							Config: c.Config{
								Flimit: 0,
							},
							Resources: c.Resources{
								Config:  config,
								Streams: streams,
								Logs:    logs,
								Errors:  errs,
							},
							Server: c.Server{
								Enabled: server,
								Open:    open,
								Host:    host,
								Port:    port,
							},
						}
						r.Blueprint.Projects = r.Blueprint.Projects[:0]
						r.Blueprint.Add(p)
					} else if len(r.Blueprint.Projects) <= 0 {
						r.Blueprint.Add(p)
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
				Name:     "add",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: r.Wdir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "/", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enables the tests"},
				},
				Action: func(p *cli.Context) (err error) {
					handle(r.Blueprint.Insert(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully added"))
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
				Usage:    "Remove a project",
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
		},
	}
	c.Run(os.Args)
}
