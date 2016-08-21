package main

import (
	"fmt"
	r "github.com/tockins/realize/realize"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
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

	header := func() error {
		app.Information()
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			log.Fatal(r.Red("$GOPATH isn't set up properly"))
		}
		return nil
	}

	wd := func() string {
		dir, err := os.Getwd()
		if err != nil {
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
			{
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
				Name:  "fast",
				Usage: "Build and watch file changes for a single project without any config file",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Usage: "Disable go run"},
					&cli.BoolFlag{Name: "bin", Usage: "Disable go install"},
					&cli.BoolFlag{Name: "fmt", Usage: "Disable gofmt"},
					&cli.BoolFlag{Name: "config", Usage: "If there is a config file with a project for the current directory take that configuration"},
				},
				Action: func(p *cli.Context) error {
					y := r.New(p)
					y.Projects[0].Path = wd()
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
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Project name \t"},
					&cli.StringFlag{Name: "path", Aliases: []string{"b"}, Value: wd(), Usage: "Project base path \t"},
					&cli.BoolFlag{Name: "build", Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Usage: "Disable go run"},
					&cli.BoolFlag{Name: "bin", Usage: "Disable go install"},
					&cli.BoolFlag{Name: "fmt", Usage: "Disable gofmt"},
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
