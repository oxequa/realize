package main

import (
	"os"
	"gopkg.in/urfave/cli.v2"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"gopkg.in/yaml.v2"
	"github.com/labstack/echo/engine/standard"
	"realize"
)

func main() {

	t := realize.Config{App_file:"realize.config.yml"}
	t.Create()

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

	// web server
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Run(standard.New("0.0.0.0:8080"))
}