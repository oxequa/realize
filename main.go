package main

import (
	"os"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"github.com/tockins/realize/realize"
)

func main() {

	handle := func(err error) error{
		if err != nil {
			return cli.Exit(err.Error(), 86)
		}
		return nil
	}

	app := &cli.App{
		Name: "Realize",
		Version: "v1.0",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Alessio Pracchia",
				Email: "pracchia@hastega.it",
			},
		},
		Usage: "A sort of Webpack for Go. Run, build and watch file changes with custom paths",
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
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(params *cli.Context) error {
					y := realize.Config{}
					y.Init(params)
					return handle(y.Create())
				},
			},
			{
				Name:     "add",
				Category: "config",
				Aliases:     []string{"s"},
				Usage: "add another project in config file",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "main", Aliases: []string{"m"}, Value: "main.go"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: true},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: true},
				},
				Action: func(params *cli.Context) error {
					y := realize.Config{}
					err := y.Read()
					return handle(err)
				},
			},
		},
		//Flags: []cli.Flag {
		//	&cli.StringFlag{
		//		Name:    "run",
		//		Value:   "main.go",
		//		Usage:   "main file of your project",
		//	},
		//},
	}

	app.Run(os.Args)
}