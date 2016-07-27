package main

import (
	"os"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"github.com/tockins/realize/realize"
)

const(
	name = "Realize"
	version = "v1.0"
	email = "pracchia@hastega.it"
	description = "Run and build your applications on file changes. Watch custom paths and specific extensions. Define additional commands and multiple projects"
	author = "Alessio Pracchia"
)

func main() {

	handle := func(err error) error{
		if err != nil {
			return cli.Exit(err.Error(), 86)
		}
		return nil
	}

	app := &cli.App{
		Name: name,
		Version: version,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  author,
				Email: email,
			},
		},
		Usage: description,
		Commands: []*cli.Command{
			{
				Name: "run",
				Usage: "Build and watch file changes",
				Action: func(p *cli.Context) error {
					fmt.Printf("Hello %q", p.String("run"))
					return nil
				},
			},
			{
				Name:     "start",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "Create the initial config",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(p *cli.Context) error {
					y := realize.New(p)
					return handle(y.Create(p))
				},
			},
			{
				Name:     "add",
				Category: "config",
				Aliases:     []string{"a"},
				Usage: "Add another project",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(p *cli.Context) error {
					y := realize.New(p)
					return handle(y.Add(p))
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
					y := realize.New(p)
					return handle(y.Remove(p))
				},
			},
			{
				Name:     "list",
				Category: "config",
				Aliases:     []string{"l"},
				Usage: "Projects list",
				Action: func(p *cli.Context) error {
					y := realize.New(p)
					return handle(y.List())
				},
			},
		},
	}

	app.Run(os.Args)
}