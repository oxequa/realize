package main

import (
	"os"
	"gopkg.in/urfave/cli.v2"
	r "github.com/tockins/realize/realize"
)

func main() {

	app := r.Init()

	handle := func(err error) error {
		if err != nil {
			r.Fail(err.Error())
			return nil
		}
		return nil
	}

	header := func() {
		app.Information()
	}

	cli := &cli.App{
		Name: app.Name,
		Version: app.Version,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  app.Author,
				Email: app.Email,
			},
		},
		Usage: app.Description,
		Commands: []*cli.Command{
			{
				Name: "run",
				Usage: "Build and watch file changes",
				Action: func(p *cli.Context) error {
					y := r.New(p)
					y.Watch()
					return nil
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "start",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "Create the initial config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App", Usage: "Project name \t"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go", Usage: "Project main file \t"},
					&cli.StringFlag{Name: "base", Aliases: []string{"b"}, Value: "/", Usage: "Project base path \t"},
					&cli.BoolFlag{Name: "build", Value: false},
					&cli.BoolFlag{Name: "run", Value: true},
					&cli.BoolFlag{Name: "bin", Value: true},
				},
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.Create(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "add",
				Category: "config",
				Aliases:     []string{"a"},
				Usage: "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App", Usage: "Project name \t"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go", Usage: "Project main file \t"},
					&cli.StringFlag{Name: "base", Aliases: []string{"b"}, Value: "/", Usage: "Project base path \t"},
					&cli.BoolFlag{Name: "build", Value: false},
					&cli.BoolFlag{Name: "run", Value: true},
					&cli.BoolFlag{Name: "bin", Value: true},
				},
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.Add(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "remove",
				Category: "config",
				Aliases:     []string{"r"},
				Usage: "Remove a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
				},
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.Remove(p))
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "list",
				Category: "config",
				Aliases:     []string{"l"},
				Usage: "Projects list",
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.List())
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