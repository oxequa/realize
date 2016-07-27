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
				Action: func(c *cli.Context) error {
					fmt.Printf("Hello %q", c.String("run"))
					return nil
				},
			},
			{
				Name:     "start",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "create the initial config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(params *cli.Context) error {
					y := realize.New(params)
					return handle(y.Create(params))
				},
			},
			{
				Name:     "add",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "add another project in config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(params *cli.Context) error {
					y := realize.New(params)
					return handle(y.Add(params))
				},
			},
			{
				Name:     "remove",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "remove a project in config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "Sample App"},
				},
				Action: func(params *cli.Context) error {
					y := realize.New(params)
					return handle(y.Remove(params))
				},
			},
			{
				Name:     "list",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "projects list",
				Action: func(params *cli.Context) error {
					y := realize.New(params)
					return handle(y.List())
				},
			},
		},
	}

	app.Run(os.Args)
}