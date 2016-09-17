package main

import (
	"fmt"
	c "github.com/tockins/realize/cli"
	s "github.com/tockins/realize/server"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"syscall"
)

var App Realize

// Realize struct contains the general app informations
type Realize struct {
	Name, Description, Author, Email string
	Version                          string
	Limit                            uint64
	Blueprint                        c.Blueprint
	Server                           s.Server
}

// Flimit defines the max number of watched files
func (r *Realize) Increases() {
	// increases the files limit
	var rLimit syscall.Rlimit
	rLimit.Max = r.Limit
	rLimit.Cur = r.Limit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println(c.Red("Error Setting Rlimit "), err)
	}
}

func init() {
	App = Realize{
		Name:        "Realize",
		Version:     "1.0",
		Description: "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Limit:       10000,
		Blueprint: c.Blueprint{
			Files: map[string]string{
				"config": "r.config.yaml",
				"output": "r.output.log",
			},
		},
	}
	App.Increases()
	c.Bp = &App.Blueprint
	s.Bp = &App.Blueprint

}

func main() {

	handle := func(err error) error {
		if err != nil {
			fmt.Println(c.Red(err.Error()))
			return nil
		}
		return nil
	}

	header := func() error {
		fmt.Println(c.Blue(App.Name) + " - " + c.Blue(App.Version))
		fmt.Println(c.BlueS(App.Description) + "\n")
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			log.Fatal(c.Red("$GOPATH isn't set up properly"))
		}
		return nil
	}

	cli := &cli.App{
		Name:    App.Name,
		Version: App.Version,
		Authors: []*cli.Author{
			{
				Name:  "Alessio Pracchia",
				Email: "pracchia@hastega.it",
			},
			{
				Name:  "Daniele Conventi",
				Email: "conventi@hastega.it",
			},
		},
		Usage: App.Description,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Build and watch file changes",
				Action: func(p *cli.Context) error {
					return handle(App.Blueprint.Run())
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:  "fast",
				Usage: "Build and watch file changes for a single project without any Configuration file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enables the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enable the tests"},
					&cli.BoolFlag{Name: "Configuration", Value: false, Usage: "Take the defined settings if exist a Configuration file."},
				},
				Action: func(p *cli.Context) error {
					App.Blueprint.Add(p)
					App.Server.Start()
					return handle(App.Blueprint.Fast(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "add",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: c.Wdir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "/", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enable the tests"},
				},
				Action: func(p *cli.Context) error {
					return handle(App.Blueprint.Insert(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
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
					return handle(App.Blueprint.Remove(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "list",
				Category: "Configuration",
				Aliases:  []string{"l"},
				Usage:    "Projects list",
				Action: func(p *cli.Context) error {
					return handle(App.Blueprint.List())
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
		},
	}

	cli.Run(os.Args)
}
