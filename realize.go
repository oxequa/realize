package main

import (
	"fmt"
	"go/build"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	RPrefix  = "realize"
	RVersion = "2.0"
	RExt     = ".yaml"
	RFile    = RPrefix + RExt
	RDir     = "." + RPrefix
	RExtWin  = ".exe"
)

type (
	Realize struct {
		Settings Settings `yaml:"settings" json:"settings"`
		Server   Server   `yaml:"server" json:"server"`
		Schema   `yaml:",inline"`
		sync     chan string
		exit     chan os.Signal
	}
	LogWriter struct{}
)

var r Realize

// init check
func init() {
	// custom log
	log.SetFlags(0)
	log.SetOutput(LogWriter{})
	if build.Default.GOPATH == "" {
		log.Fatal("$GOPATH isn't set properly")
	}
	if err := os.Setenv("GOBIN", filepath.Join(build.Default.GOPATH, "bin")); err != nil {
		log.Fatal(err)
	}
}

// Realize cli commands
func main() {
	app := &cli.App{
		Name:        strings.Title(RPrefix),
		Version:     RVersion,
		Description: "Realize is the #1 Golang Task Runner which enhance your workflow by automating the most common tasks and using the best performing Golang live reloading.",
		Commands: []*cli.Command{
			{
				Name:        "start",
				Aliases:     []string{"s"},
				Description: "Start " + strings.Title(RPrefix) + " on a given path. If not exist a config file it creates a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: ".", Usage: "Project base path"},
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "", Usage: "Run a project by its name"},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "vet", Aliases: []string{"v"}, Value: false, Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "server", Aliases: []string{"s"}, Value: false, Usage: "Enable server and open into the default browser"},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing config and doesn't create a new one"},
				},
				Action: func(c *cli.Context) error {
					return r.start(c)
				},
			},
			{
				Name:        "add",
				Category:    "Configuration",
				Aliases:     []string{"a"},
				Description: "Add a project to an existing config or to a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: wdir(), Usage: "Project base path"},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "vet", Aliases: []string{"v"}, Value: false, Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
				},
				Action: func(c *cli.Context) error {
					return r.add(c)
				},
			},
			{
				Name:        "init",
				Category:    "Configuration",
				Aliases:     []string{"i"},
				Description: "Make a new config file step by step.",
				Action: func(c *cli.Context) error {
					return r.setup(c)
				},
			},
			{
				Name:        "remove",
				Category:    "Configuration",
				Aliases:     []string{"r"},
				Description: "Remove a project from an existing config.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
				},
				Action: func(c *cli.Context) error {
					return r.remove(c)
				},
			},
			{
				Name:        "clean",
				Category:    "Configuration",
				Aliases:     []string{"c"},
				Description: "Remove " + strings.Title(RPrefix) + " folder.",
				Action: func(c *cli.Context) error {
					return r.clean()
				},
			},
			{
				Name:        "version",
				Aliases:     []string{"v"},
				Description: "Print " + strings.Title(RPrefix) + " version.",
				Action: func(p *cli.Context) error {
					r.version()
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// Stop realize workflow
func (r *Realize) Stop() {
	close(r.exit)
}

// Run realize workflow
func (r *Realize) Start() {
	r.exit = make(chan os.Signal, 2)
	signal.Notify(r.exit, os.Interrupt, syscall.SIGTERM)
	for k := range r.Schema.Projects {
		r.Schema.Projects[k].parent = r
		r.Schema.Projects[k].Setup()
		go r.Schema.Projects[k].Watch(r.exit)
	}
	for {
		select {
		case <-r.exit:
			return
		}
	}
}

// Prefix a given string with tool name
func (r *Realize) Prefix(input string) string {
	if len(input) > 0 {
		return fmt.Sprint(yellow.bold("["), strings.ToUpper(RPrefix), yellow.bold("]"), " : ", input)
	}
	return input
}

// Rewrite the layout of the log timestamp
func (w LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Fprint(output, yellow.regular("["), time.Now().Format("15:04:05"), yellow.regular("]"), string(bytes))
}
