package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/tockins/interact"
	"github.com/tockins/realize/server"
	"github.com/tockins/realize/settings"
	"github.com/tockins/realize/watcher"
	cli "gopkg.in/urfave/cli.v2"
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
	settings.Settings                               `yaml:"settings,omitempty"`
	Name, Description, Author, Email, Host, Version string             `yaml:"-"`
	Sync                                            chan string        `yaml:"-"`
	Blueprint                                       watcher.Blueprint  `yaml:"-"`
	Server                                          server.Server      `yaml:"-"`
	Projects                                        *[]watcher.Project `yaml:"projects" json:"projects"`
}

// Realize struct initialization
func init() {
	r = realize{
		Name:        name,
		Version:     version,
		Description: description,
		Sync:        make(chan string),
		Settings: settings.Settings{
			Config: settings.Config{
				Create: true,
			},
			Resources: settings.Resources{
				Config:  config,
				Outputs: outputs,
				Logs:    logs,
				Errors:  errs,
			},
			Server: settings.Server{
				Status: false,
				Open:   false,
				Host:   host,
				Port:   port,
			},
		},
	}
	r.Blueprint = watcher.Blueprint{
		Settings: &r.Settings,
		Sync:     r.Sync,
	}
	r.Server = server.Server{
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
				Email: "pracchia@hastega.it",
			},
			{
				Name:  "Daniele Conventi",
				Email: "conventi@hastega.it",
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
					&cli.BoolFlag{Name: "no-install", Aliases: []string{"ni"}, Value: false, Usage: "Disable go install"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing configurations."},
				},
				Action: func(p *cli.Context) error {
					if p.Bool("legacy") {
						r.Config.Legacy = settings.Legacy{
							Status:   p.Bool("legacy"),
							Interval: interval,
						}
					}
					if p.Bool("no-config") || len(r.Blueprint.Projects) <= 0 {
						if p.Bool("no-config") {
							r.Config.Create = false
						}
						r.Blueprint.Projects = []watcher.Project{}
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
					fmt.Println(r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"), r.Green.Bold("Your project was successfully added."))
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
					interact.Run(&interact.Interact{
						Before: func(context interact.Context) error {
							context.SetErr(r.Red.Bold("INVALID INPUT"))
							context.SetPrfx(color.Output, r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"))
							return nil
						},
						Questions: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									if _, err := os.Stat(".realize/" + config); err != nil {
										d.Skip()
									}
									d.SetDef(false, r.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: r.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to overwrite the existing " + r.Colors.Magenta.Bold("Realize") + " config?",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									} else if val {
										r.Settings = settings.Settings{
											Config: settings.Config{
												Create: true,
											},
											Resources: settings.Resources{
												Config:  config,
												Outputs: outputs,
												Logs:    logs,
												Errors:  errs,
											},
											Server: settings.Server{
												Status: false,
												Open:   false,
												Host:   host,
												Port:   port,
											},
										}
										r.Blueprint.Projects = r.Blueprint.Projects[len(r.Blueprint.Projects):]
									}
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, r.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: r.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to customize the " + r.Colors.Magenta.Bold("settings") + "?",
									Resolve: func(d interact.Context) bool {
										val, _ := d.Ans().Bool()
										return val
									},
								},
								Subs: []*interact.Question{
									{
										Before: func(d interact.Context) error {
											d.SetDef(0, r.Green.Regular("(os default)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[int]"),
											Msg:     "Max number of open files (root required)",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Int()
											if err != nil {
												return d.Err()
											}
											r.Config.Flimit = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable legacy watch by polling",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(1, r.Green.Regular("(1s)"))
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[seconds]"),
													Msg:     "Set polling interval in seconds",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Int()
													if err != nil {
														return d.Err()
													}
													r.Config.Legacy.Interval = time.Duration(val * 1000000000)
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Config.Legacy.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable web server",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(5001, r.Green.Regular("(5001)"))
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[int]"),
													Msg:     "Server port",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Int()
													if err != nil {
														return d.Err()
													}
													r.Server.Port = int(val)
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef("localhost", r.Green.Regular("(localhost)"))
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Server host",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Server.Host = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, r.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[y/n]"),
													Msg:     "Open in the current browser",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Server.Open = val
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Server.Status = val
											return nil
										},
									},
								},
								Action: func(d interact.Context) interface{} {
									_, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									}
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(true, r.Green.Regular("(y)"))
									d.SetEnd("!")
									return nil
								},
								Quest: interact.Quest{
									Options: r.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to " + r.Colors.Magenta.Regular("add a new project") + "? (insert '!' to stop)",
									Resolve: func(d interact.Context) bool {
										val, _ := d.Ans().Bool()
										if val {
											r.Blueprint.Add(p)
										}
										return val
									},
								},
								Subs: []*interact.Question{
									{
										Before: func(d interact.Context) error {
											d.SetDef(r.Settings.Wdir(), r.Green.Regular("("+r.Settings.Wdir()+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[string]"),
											Msg:     "Project name",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Name = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											dir, _ := os.Getwd()
											d.SetDef(dir, r.Green.Regular("("+dir+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[string]"),
											Msg:     "Project path",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Path = r.Settings.Path(val)
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, r.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go fmt",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Fmt = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go test",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Test = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go generate",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Generate = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, r.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go install",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Bin = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go build",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Build = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, r.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable go run",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Run = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Customize the watched paths",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												if val {
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Paths = r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Paths[:len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Paths)-1]
												}
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetEnd("!")
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a path to watch (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Paths = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Paths, val)
													d.Reload()
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											_, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Customize the ignored paths",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												if val {
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Ignore = r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Ignore[:len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Ignore)-1]
												}
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetEnd("!")
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a path to ignore (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Ignore = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Ignore, val)
													d.Reload()
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											_, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Add additionals arguments",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetEnd("!")
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert an argument (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Params = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Params, val)
													d.Reload()
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											_, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Add 'before' custom commands",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetEnd("!")
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a command (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts, watcher.Command{Type: "before", Command: val})
													d.Reload()
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											_, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Add 'after' custom commands",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetEnd("!")
													return nil
												},
												Quest: interact.Quest{
													Options: r.Yellow.Regular("[string]"),
													Msg:     "Insert a command (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts, watcher.Command{Type: "after", Command: val})
													d.Reload()
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											_, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable watcher files preview",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Preview = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable file output history",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Streams.FileOut = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable file logs history",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Streams.FileLog = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, r.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: r.Yellow.Regular("[y/n]"),
											Msg:     "Enable file errors history",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Streams.FileErr = val
											return nil
										},
									},
								},
								Action: func(d interact.Context) interface{} {
									if val, err := d.Ans().Bool(); err != nil {
										return d.Err()
									} else if val {
										d.Reload()
									}
									return nil
								},
							},
						},
						After: func(d interact.Context) error {
							if val, _ := d.Qns().Get(0).Ans().Bool(); val {
								err = r.Settings.Remove()
								if err != nil {
									return err
								}
							}
							return nil
						},
					})
					handle(r.Record(r))
					fmt.Println(r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"), r.Green.Bold("Your configuration was successful."))
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
					fmt.Println(r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"), r.Green.Bold("Your project was successfully removed."))
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
					fmt.Println(r.Yellow.Bold("[")+"REALIZE"+r.Yellow.Bold("]"), r.Green.Bold("Realize folder successfully removed."))
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
