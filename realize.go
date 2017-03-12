package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	i "github.com/tockins/interact"
	s "github.com/tockins/realize/server"
	c "github.com/tockins/realize/settings"
	w "github.com/tockins/realize/watcher"
	"gopkg.in/urfave/cli.v2"
	"os"
	"time"
)

const (
	name        = "Realize"
	version     = "1.3"
	description = "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths"
	config      = "realize.yaml"
	outputs     = "outputs.log"
	errs        = "errors.log"
	logs        = "logs.log"
	host        = "localhost"
	port        = 5001
	interval    = 200
)

var r realize

// Realize struct contains the general app informations
type realize struct {
	c.Settings                                      `yaml:"settings,omitempty"`
	Name, Description, Author, Email, Host, Version string       `yaml:"-"`
	Sync                                            chan string  `yaml:"-"`
	Blueprint                                       w.Blueprint  `yaml:"-"`
	Server                                          s.Server     `yaml:"-"`
	Projects                                        *[]w.Project `yaml:"projects" json:"projects"`
}

// Realize struct initialization
func init() {
	r = realize{
		Name:        name,
		Version:     version,
		Description: description,
		Sync:        make(chan string),
		Settings: c.Settings{
			Config: c.Config{
				Create: true,
			},
			Resources: c.Resources{
				Config:  config,
				Outputs: outputs,
				Logs:    logs,
				Errors:  errs,
			},
			Server: c.Server{
				Status: true,
				Open:   false,
				Host:   host,
				Port:   port,
			},
		},
	}
	r.Blueprint = w.Blueprint{
		Settings: &r.Settings,
		Sync:     r.Sync,
	}
	r.Server = s.Server{
		Blueprint: &r.Blueprint,
		Settings:  &r.Settings,
		Sync:      r.Sync,
	}
	r.Projects = &r.Blueprint.Projects

	// read if exist
	r.Read(&r)

	// increase the file limit
	if r.Config.Flimit != 0 {
		r.Flimit()
	}
}

// Before of every exec of a cli method
func before() error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return handle(errors.New("$GOPATH isn't set up properly"))
	}
	return nil
}

// Handle errors
func handle(err error) error {
	if err != nil {
		fmt.Println(r.Red.Bold(err.Error()))
		os.Exit(1)
	}
	return nil
}

// Cli commands
func main() {
	app := &cli.App{
		Name:    r.Name,
		Version: r.Version,
		Authors: []*cli.Author{
			{
				Name:  "Alessio Pracchia",
				Email: "pracchia@hastegit",
			},
			{
				Name:  "Daniele Conventi",
				Email: "conventi@hastegit",
			},
		},
		Usage: r.Description,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run a toolchain on a project or a list of projects. If not exist a config file it creates a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path."},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test."},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate."},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build."},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Watch by polling instead of Watch by fsnotify."},
					&cli.BoolFlag{Name: "server", Aliases: []string{"s"}, Value: false, Usage: "Enable server and open into the default browser."},
					&cli.BoolFlag{Name: "no-run", Aliases: []string{"nr"}, Value: false, Usage: "Disable go run"},
					&cli.BoolFlag{Name: "no-fmt", Aliases: []string{"nf"}, Value: false, Usage: "Disable go fmt."},
					&cli.BoolFlag{Name: "no-install", Aliases: []string{"ni"}, Value: false, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing configurations."},
				},
				Action: func(p *cli.Context) error {
					if p.Bool("legacy") {
						r.Config.Legacy = c.Legacy{
							Status:   p.Bool("legacy"),
							Interval: interval,
						}
					}
					if p.Bool("no-config") || len(r.Blueprint.Projects) <= 0 {
						if p.Bool("no-config") {
							r.Config.Create = false
						}
						r.Blueprint.Projects = []w.Project{}
						handle(r.Blueprint.Add(p))
					}
					handle(r.Server.Start(p))
					handle(r.Blueprint.Run())
					handle(r.Record(r))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "add",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Add a project to an existing config file or create a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path."},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test."},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate."},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build."},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Watch by polling instead of Watch by fsnotify."},
					&cli.BoolFlag{Name: "server", Aliases: []string{"s"}, Value: false, Usage: "Enable server and open into the default browser."},
					&cli.BoolFlag{Name: "no-run", Aliases: []string{"nr"}, Value: false, Usage: "Disable go run"},
					&cli.BoolFlag{Name: "no-fmt", Aliases: []string{"nf"}, Value: false, Usage: "Disable go fmt."},
					&cli.BoolFlag{Name: "no-install", Aliases: []string{"ni"}, Value: false, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing configurations."},
				},
				Action: func(p *cli.Context) (err error) {
					fmt.Println(p.String("path"))
					handle(r.Blueprint.Add(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully added."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "init",
				Category: "Configuration",
				Aliases:  []string{"a"},
				Usage:    "Define a new config file with all options step by step",
				Action: func(p *cli.Context) (err error) {
					i.Run(&i.Interact{
						Before: func(context i.Context) error {
							r.Blueprint.Add(p)
							context.SetErr(r.Red.Bold("INVALID INPUT"))
							context.SetPrfx(color.Output, r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"))
							return nil
						},
						Questions: []*i.Question{
							{
								Before: func(c i.Context) error {
									if _, err := os.Stat(".realize/" + config); err != nil {
										c.Skip()
									}
									return nil
								},
								Quest: i.Quest{
									Options: r.Yellow.Regular("[y/n]"),
									Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
									Msg:     "Would you want overwrite the existing Realize config?",
								},
								Action: func(c i.Context) interface{} {
									val, err := c.Ans().Bool()
									if err != nil {
										return c.Err()
									} else if val {
										err = r.Settings.Remove()
										if err != nil {
											return err
										}
									}
									return nil
								},
							},
							{
								Quest: i.Quest{
									Options: r.Yellow.Regular("[y/n]"),
									Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
									Msg:     "Would you want customize the general settings?",
									Resolve: func(c i.Context) bool {
										val, _ := c.Ans().Bool()
										if val {
											r.Blueprint.Add(p)
										}
										return val
									},
								},
								Subs: []*i.Question{
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[int]"),
											Default: i.Default{Value: 0, Preview: true, Text: r.Green.Regular("(os default)")},
											Msg:     "Max number of open files (root required)",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Int()
											if err != nil {
												return c.Err()
											}
											r.Config.Flimit = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable legacy watch by polling",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[int]"),
													Default: i.Default{Value: 1, Preview: true, Text: r.Green.Regular("(1s)")},
													Msg:     "Set polling interval in seconds",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().Int()
													if err != nil {
														return c.Err()
													}
													r.Config.Legacy.Interval = time.Duration(val * 1000)
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Config.Legacy.Status = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[string]"),
											Default: i.Default{Value: r.Settings.Wdir(), Preview: true, Text: r.Green.Regular("(" + r.Settings.Wdir() + ")")},
											Msg:     "Project name",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().String()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Name = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[string]"),
											Default: i.Default{Value: "", Preview: true, Text: r.Green.Regular("(current wdir)")},
											Msg:     "Project path",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().String()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Path = r.Settings.Path(val)
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: true, Preview: true, Text: r.Green.Regular("(y)")},
											Msg:     "Enable go fmt",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Fmt = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable go test",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Test = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable go generate",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Generate = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: true, Preview: true, Text: r.Green.Regular("(y)")},
											Msg:     "Enable go install",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Bin = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable go build",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Build = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable go run",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Run = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Would you want customize the watched paths?",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												if val {
													r.Blueprint.Projects[0].Watcher.Paths = r.Blueprint.Projects[0].Watcher.Paths[:len(r.Blueprint.Projects[0].Watcher.Paths)-1]
												}
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a path to watch (insert '!' to stop)",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													if val == "!" {
														return nil
													} else {
														r.Blueprint.Projects[0].Watcher.Paths = append(r.Blueprint.Projects[0].Watcher.Paths, val)
														c.Reload()
													}
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											_, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Would you want customize the ignored paths?",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												if val {
													r.Blueprint.Projects[0].Watcher.Ignore = r.Blueprint.Projects[0].Watcher.Ignore[:len(r.Blueprint.Projects[0].Watcher.Ignore)-1]
												}
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a path to ignore (insert '!' to stop)",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													if val == "!" {
														return nil
													} else {
														r.Blueprint.Projects[0].Watcher.Ignore = append(r.Blueprint.Projects[0].Watcher.Ignore, val)
														c.Reload()
													}
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											_, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Would you want add additional arguments?",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert an argument (insert '!' to stop)",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													if val == "!" {
														return nil
													} else {
														r.Blueprint.Projects[0].Params = append(r.Blueprint.Projects[0].Params, val)
														c.Reload()
													}
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											_, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Would you want add 'before' custom commands?",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a command (insert '!' to stop)",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													if val == "!" {
														return nil
													} else {
														r.Blueprint.Projects[0].Watcher.Scripts = append(r.Blueprint.Projects[0].Watcher.Scripts, w.Command{Type: "before", Command: val})
														c.Reload()
													}
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											_, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Would you want add 'after' custom commands?",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a command (insert '!' to stop)",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													if val == "!" {
														return nil
													} else {
														r.Blueprint.Projects[0].Watcher.Scripts = append(r.Blueprint.Projects[0].Watcher.Scripts, w.Command{Type: "after", Command: val})
														c.Reload()
													}
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											_, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable watcher files preview",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Run = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable web server",
											Resolve: func(c i.Context) bool {
												val, _ := c.Ans().Bool()
												return val
											},
										},
										Subs: []*i.Question{
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[int]"),
													Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(5001)")},
													Msg:     "Server port",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().Int()
													if err != nil {
														return c.Err()
													}
													r.Server.Port = int(val)
													return nil
												},
											},
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[string]"),
													Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(localhost)")},
													Msg:     "Server host",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().String()
													if err != nil {
														return c.Err()
													}
													r.Server.Host = val
													return nil
												},
											},
											{
												Quest: i.Quest{
													Options: r.Yellow.Regular("[y/n]"),
													Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
													Msg:     "Open in the current browser",
												},
												Action: func(c i.Context) interface{} {
													val, err := c.Ans().Bool()
													if err != nil {
														return c.Err()
													}
													r.Server.Open = val
													return nil
												},
											},
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Server.Status = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable file output history",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Streams.FileOut = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable file logs history",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Streams.FileLog = val
											return nil
										},
									},
									{
										Quest: i.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Default: i.Default{Value: false, Preview: true, Text: r.Green.Regular("(n)")},
											Msg:     "Enable file errors history",
										},
										Action: func(c i.Context) interface{} {
											val, err := c.Ans().Bool()
											if err != nil {
												return c.Err()
											}
											r.Blueprint.Projects[0].Streams.FileErr = val
											return nil
										},
									},
								},
								Action: func(c i.Context) interface{} {
									if _, err := c.Ans().Bool(); err != nil {
										return c.Err()
									}
									return nil
								},
							},
						},
					})
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully added."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "remove",
				Category: "Configuration",
				Aliases:  []string{"r"},
				Usage:    "Remove a project from a realize configuration.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
				},
				Action: func(p *cli.Context) error {
					handle(r.Blueprint.Remove(p))
					handle(r.Record(r))
					fmt.Println(r.Green.Bold("Your project was successfully removed."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "list",
				Category: "Configuration",
				Aliases:  []string{"l"},
				Usage:    "Print projects list.",
				Action: func(p *cli.Context) error {
					return handle(r.Blueprint.List())
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
			{
				Name:     "clean",
				Category: "Configuration",
				Aliases:  []string{"c"},
				Usage:    "Remove realize folder.",
				Action: func(p *cli.Context) error {
					handle(r.Settings.Remove())
					fmt.Println(r.Green.Bold("Realize folder successfully removed."))
					return nil
				},
				Before: func(c *cli.Context) error {
					return before()
				},
			},
		},
	}
	app.Run(os.Args)
}
