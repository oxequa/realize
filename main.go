package main

import (
	//"os"
	//"gopkg.in/urfave/cli.v2"
	"github.com/tockins/realize/realize"
	"gopkg.in/urfave/cli.v2"
	"fmt"
	"os"
)

func main() {

	app := &cli.App{
		Name: "realize",
		Version: "1.0",
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
				Usage: "create the initial config file",
				Action: func(c *cli.Context) error {
					realize.Init().Create()
					fmt.Printf("Hello %q", c.String("run"))
					return nil
				},
			},
		},
		Flags: []cli.Flag {
			&cli.StringFlag{
				Name:    "run",
				Value:   "main.go",
				Usage:   "main file of your project",
			},
		},
	}

	app.Run(os.Args)
}