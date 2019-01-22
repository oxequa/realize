package main

import (
	"github.com/oxequa/realize3"
	"github.com/urfave/cli"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	name        = "Realize"
	version     = "3.0"
	description = "Go task runner and watcher, automate painful and time-consuming tasks in your development workflow"
)

func main() {
	// custom logs
	log.SetFlags(0)
	log.SetOutput(core.Log{})
	// cli
	app := &cli.App{
		Name:        name,
		Version:     version,
		Description: description,
		Commands: []cli.Command{
			{
				Name:        "start",
				Aliases:     []string{"s"},
				Description: "Start on a given Go project. Create a config if it doesn't already exist.",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "fmt", Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "mod", Usage: "Enable go mod"},
					&cli.BoolFlag{Name: "vet", Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "run", Usage: "Enable go run"},
					&cli.BoolFlag{Name: "test", Usage: "Enable go test"},
					&cli.BoolFlag{Name: "build", Usage: "Enable go build"},
					&cli.BoolFlag{Name: "panel", Usage: "Start web panel"},
					&cli.BoolFlag{Name: "install", Usage: "Enable go install"},
					&cli.BoolFlag{Name: "generate", Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "polling", Usage: "Enable watch by polling"},
					&cli.BoolFlag{Name: "raw", Usage: "Start without reading/making a config"},
					&cli.StringFlag{Name: "path", Value: ".", Usage: "Start on custom path"},
					&cli.StringFlag{Name: "name", Value: "", Usage: "Start filtering by project name"},
				},
				Action: start,
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(core.Prefix("Error", core.Red), err)
	}
}

func start(c *cli.Context) error {
	r := core.Realize{}
	// Read/write config file
	if !c.Bool("raw") {
		r.Settings.Read(&r)
	}
	// Check polling flag
	if c.Bool("polling") {
		r.Settings.Polling.Active = c.Bool("legacy")
		r.Settings.Polling.Interval = time.Second * 1
	}
	// File limit
	if r.Settings.FileLimit != 0 {
		if err := r.Settings.Flimit(); err != nil {
			return err
		}
	}
	// Default config
	if len(r.Projects) == 0 {
		wdir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		// create project
		project := core.Project{
			Realize: &r,
			Name:    filepath.Base(wdir),
		}
		// Go mod
		if c.Bool("mod") {
			project.TasksBefore = append(project.Tasks, core.Command{Log: true, Task: "go mod init"})
			project.TasksAfter = append(project.Tasks, core.Command{Log: true, Task: "go mod tidy"})
		}
		// Go fmt
		if c.Bool("ftm") {
			project.TasksAfter = append(project.Tasks, core.Command{Log: true, Task: "go fmt ./.."})
		}
		// Go test
		if c.Bool("test") {
			project.TasksBefore = append(project.Tasks, core.Command{Log: true, Task: "go test ./.."})
		}
		// Go vet
		if c.Bool("vet") {
			project.TasksBefore = append(project.Tasks, core.Command{Log: true, Task: "go vet"})
		}
		// Go install
		if c.Bool("install") {
			project.Tasks = append(project.Tasks, core.Command{Log: true, Task: "go install"})
		}
		// Go build
		if c.Bool("build") {
			project.Tasks = append(project.Tasks, core.Command{Log: true, Task: "go build"})
		}
		// Go run
		if c.Bool("run") {
			if !c.Bool("install") {
				project.Tasks = append(project.Tasks, core.Series{
					Tasks: core.ToInterface([]core.Command{
						{
							Task: "go install",
						}, {
							Task: "go run",
						},
					}),
				})
			} else {
				project.Tasks = append(project.Tasks, core.Command{Log: true, Task: "go run"})
			}
		}
		// add project
		r.Projects = append(r.Projects, project)
	}
	// Start tasks

	// Write config
	if !c.Bool("raw") {
		err := r.Settings.Write(r)
		if err != nil {
			print(err.Error())
			return err
		}
	}
	return nil
}
