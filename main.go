package main

import (
	r "github.com/tockins/realize/realize"
	"gopkg.in/urfave/cli.v2"
	"os"
	"fmt"
	"strings"
)

func main() {

	app := r.Init()

	handle := func(err error) error {
		if err != nil {
			fmt.Println(r.Red(err.Error()))
			return nil
		}
		return nil
	}

	header := func() {
		app.Information()
	}

	wd := func() string{
		dir, err :=os.Getwd()
		if err != nil{
			fmt.Println(r.Red(err))
			return "/"
		}
		wd := strings.Split(dir, "/")
		return wd[len(wd)-1]
	}

	cli := &cli.App{
		Name:    app.Name,
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
				Name:  "run",
				Usage: "Build and watch file changes",
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.Watch())
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:     "start",
				Category: "config",
				Aliases:  []string{"s"},
				Usage:    "Create the initial config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "", Usage: "Project name \t"},
					&cli.StringFlag{Name: "base", Aliases: []string{"b"}, Value: wd(), Usage: "Project base path \t"},
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
				Aliases:  []string{"a"},
				Usage:    "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Project name \t"},
					&cli.StringFlag{Name: "base", Aliases: []string{"b"}, Value: wd(), Usage: "Project base path \t"},
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
				Aliases:  []string{"r"},
				Usage:    "Remove a project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
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
				Aliases:  []string{"l"},
				Usage:    "Projects list",
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
