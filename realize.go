package main

import (
	r "github.com/tockins/realize/cli"
	//_ "github.com/tockins/realize/server"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
)

func init() {
	App := r.Realize{
		Name:        "Realize",
		Version:     "1.0",
		Description: "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Limit:       10000,
		Blueprint: r.Blueprint{
			Files: map[string]string{
				"config": "r.config.yaml",
				"output": "r.output.log",
			},
		},
	}
	App.Increases()
	r.App = App
}

func main() {

	app := r.App

	handle := func(err error) error {
		if err != nil {
			fmt.Println(r.Red(err.Error()))
			return nil
		}
		return nil
	}

	header := func() error {
		fmt.Println(r.Blue(app.Name) + " - " + r.Blue(app.Version))
		fmt.Println(r.BlueS(app.Description) + "\n")
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			log.Fatal(r.Red("$GOPATH isn't set up properly"))
		}
		return nil
	}

	cli := &cli.App{
		Name:    app.Name,
		Version: app.Version,
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
		Usage: app.Description,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Build and watch file changes",
				Action: func(p *cli.Context) error {
					return handle(app.Blueprint.Run())
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
					app.Blueprint.Add(p)
					return handle(app.Blueprint.Fast(p))
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
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: app.Wdir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "/", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enable the tests"},
				},
				Action: func(p *cli.Context) error {
					return handle(app.Blueprint.Insert(p))
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
					return handle(app.Blueprint.Remove(p))
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
					return handle(app.Blueprint.List())
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
