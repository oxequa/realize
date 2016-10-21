package main

import (
	a "github.com/tockins/realize/app"
	"gopkg.in/urfave/cli.v2"
	"os"
)

var app a.Realize

func main() {
	c := &cli.App{
		Name:    a.Name,
		Version: a.Version,
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
		Usage: a.Description,
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Build and watch file changes",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "no-server", Usage: "Enable the web panel"},
					&cli.BoolFlag{Name: "open", Usage: "Automatically opens the web panel"},
				},
				Action: func(p *cli.Context) error {
					return app.Handle(app.Run(p))
				},
				Before: func(c *cli.Context) error {
					return app.Before(c)
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
					&cli.BoolFlag{Name: "no-server", Usage: "Disables the web panel"},
					&cli.BoolFlag{Name: "open", Usage: "Automatically opens the web panel"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enables the tests"},
					&cli.BoolFlag{Name: "config", Value: false, Usage: "Take the defined settings if exist a Configuration file."},
				},
				Action: func(p *cli.Context) error {
					return app.Handle(app.Fast(p))
				},
				Before: func(c *cli.Context) error {
					return app.Before(c)
				},
			},
			{
				Name:     "add",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: app.Dir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "/", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "test", Value: false, Usage: "Enables the tests"},
				},
				Action: func(p *cli.Context) error {
					return app.Handle(app.Add(p))
				},
				Before: func(c *cli.Context) error {
					return app.Before(c)
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
					return app.Handle(app.Remove(p))
				},
				Before: func(c *cli.Context) error {
					return app.Before(c)
				},
			},
			{
				Name:     "list",
				Category: "Configuration",
				Aliases:  []string{"l"},
				Usage:    "Projects list",
				Action: func(p *cli.Context) error {
					return app.Handle(app.List(p))
				},
				Before: func(c *cli.Context) error {
					return app.Before(c)
				},
			},
		},
	}
	c.Run(os.Args)
}
