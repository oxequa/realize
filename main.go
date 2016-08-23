package main

import (
	"fmt"
	r "github.com/tockins/realize/realize"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
)

func main() {

	app := r.Info()

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
					y := r.New(p)
					return handle(y.Watch())
				},
				Before: func(c *cli.Context) error {
					header()
					return nil
				},
			},
			{
				Name:  "fast",
				Usage: "Build and watch file changes for a single project without any config file",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enables the build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
					&cli.BoolFlag{Name: "config", Value: false, Usage: "Take the defined settings if exist a config file."},
				},
				Action: func(p *cli.Context) error {
					y := r.New(p)
					return handle(y.Fast(p))
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
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: r.WorkingDir(), Usage: "Project name"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: "/", Usage: "Project base path"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "no-run", Usage: "Disables the run"},
					&cli.BoolFlag{Name: "no-bin", Usage: "Disables the installation"},
					&cli.BoolFlag{Name: "no-fmt", Usage: "Disables the fmt (go fmt)"},
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
