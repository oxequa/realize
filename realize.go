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
	"github.com/tockins/realize/style"
	"github.com/tockins/realize/watcher"
	cli "gopkg.in/urfave/cli.v2"
	"strconv"
)

const (
	version = "1.4.1"
)

// Realize struct contains the general app informations
type realize struct {
	settings.Settings `yaml:"settings,omitempty"`
	Sync              chan string        `yaml:"-"`
	Blueprint         watcher.Blueprint  `yaml:"-"`
	Server            server.Server      `yaml:"-"`
	Projects          *[]watcher.Project `yaml:"projects" json:"projects"`
}

// New realize instance
var r realize

// Cli commands
func main() {
	app := &cli.App{
		Name:    "Realize",
		Version: version,
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
		Description: "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Commands: []*cli.Command{
			{
				Name:        "run",
				Aliases:     []string{"r"},
				Description: "Run a toolchain on a project or a list of projects. If not exist a config file it creates a new one",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path."},
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "", Usage: "Run a project by its name."},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test."},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt."},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate."},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Watch by polling instead of watch by fsnotify."},
					&cli.BoolFlag{Name: "server", Aliases: []string{"s"}, Value: false, Usage: "Enable server and open into the default browser."},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install."},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build."},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing configurations."},
				},
				Action: func(p *cli.Context) error {
					polling(p, &r.Legacy)
					if err := insert(p, &r.Blueprint); err != nil {
						return err
					}
					if !p.Bool("no-config") {
						if err := r.Record(r); err != nil {
							return err
						}
					}
					if err := r.Server.Start(p); err != nil {
						return err
					}
					if err := r.Blueprint.Run(p); err != nil {
						return err
					}
					return nil
				},
				Before: before,
			},
			{
				Name:        "add",
				Category:    "Configuration",
				Aliases:     []string{"a"},
				Description: "Add a project to an existing config file or create a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path."},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test."},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate."},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Watch by polling instead of Watch by fsnotify."},
					&cli.BoolFlag{Name: "server", Aliases: []string{"s"}, Value: false, Usage: "Enable server and open into the default browser."},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"r"}, Value: false, Usage: "Enable go run"},
				},
				Action: func(p *cli.Context) error {
					if err := r.Blueprint.Add(p); err != nil {
						return err
					}
					if err := r.Record(r); err != nil {
						return err
					}
					fmt.Fprintln(style.Output, prefix(style.Green.Bold("Your project was successfully added.")))
					return nil
				},
				Before: before,
			},
			{
				Name:        "init",
				Category:    "Configuration",
				Aliases:     []string{"a"},
				Description: "Define a new config file with all options step by step",
				Action: func(p *cli.Context) (actErr error) {
					interact.Run(&interact.Interact{
						Before: func(context interact.Context) error {
							context.SetErr(style.Red.Bold("INVALID INPUT"))
							context.SetPrfx(color.Output, style.Yellow.Bold("[")+"REALIZE"+style.Yellow.Bold("]"))
							return nil
						},
						Questions: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									if _, err := os.Stat(settings.Directory + "/" + settings.File); err != nil {
										d.Skip()
									}
									d.SetDef(false, style.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: style.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to overwrite the existing " + style.Magenta.Bold("Realize") + " config?",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									} else if val {
										r.Settings = settings.Settings{
											File: settings.File,
											Server: settings.Server{
												Status: false,
												Open:   false,
												Host:   server.Host,
												Port:   server.Port,
											},
										}
										r.Blueprint.Projects = r.Blueprint.Projects[len(r.Blueprint.Projects):]
									}
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, style.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: style.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to customize the " + ("settings") + "?",
									Resolve: func(d interact.Context) bool {
										val, _ := d.Ans().Bool()
										return val
									},
								},
								Subs: []*interact.Question{
									{
										Before: func(d interact.Context) error {
											d.SetDef(0, style.Green.Regular("(os default)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[int]"),
											Msg:     "Max number of open files (root required)",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Int()
											if err != nil {
												return d.Err()
											}
											r.FileLimit = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable legacy watch by polling",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(1, style.Green.Regular("(1s)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[seconds]"),
													Msg:     "Set polling interval in seconds",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Int()
													if err != nil {
														return d.Err()
													}
													r.Legacy.Interval = time.Duration(val * 1000000000)
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable logging files",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Files.Errors = settings.Resource{Name: settings.FileErr, Status: val}
											r.Files.Outputs = settings.Resource{Name: settings.FileOut, Status: val}
											r.Files.Logs = settings.Resource{Name: settings.FileLog, Status: val}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable web server",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(server.Port, style.Green.Regular("("+strconv.Itoa(server.Port)+")"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[int]"),
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
													d.SetDef(server.Host, style.Green.Regular("("+server.Host+")"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
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
													d.SetDef(false, style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[y/n]"),
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
									d.SetDef(true, style.Green.Regular("(y)"))
									d.SetEnd("!")
									return nil
								},
								Quest: interact.Quest{
									Options: style.Yellow.Regular("[y/n]"),
									Msg:     "Would you want to " + style.Magenta.Regular("add a new project") + "? (insert '!' to stop)",
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
											d.SetDef(r.Settings.Wdir(), style.Green.Regular("("+r.Settings.Wdir()+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[string]"),
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
											d.SetDef(dir, style.Green.Regular("("+dir+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[string]"),
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
											d.SetDef(true, style.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go vet",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Vet = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, style.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go fmt",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Fmt additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Fmt.Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Fmt.Args, val)
													}
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Fmt.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, style.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go test",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Test additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Test.Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Test.Args, val)
													}
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Test.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go generate",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Generate additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Generate.Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Generate.Args, val)
													}
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Generate.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, style.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go install",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Install additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Install.Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Install.Args, val)
													}
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Install.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go build",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Build additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Build.Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Build.Args, val)
													}
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Build.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, style.Green.Regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Enable go run",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Cmds.Run = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Customize watched paths",
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
													Options: style.Yellow.Regular("[string]"),
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
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Customize ignored paths",
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
													Options: style.Yellow.Regular("[string]"),
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
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Add an additional argument",
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
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Add another argument (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Args = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Args, val)
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
											d.SetDef(false, style.Green.Regular("(none)"))
											d.SetEnd("!")
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Add a 'before' custom command (insert '!' to stop)",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Insert a command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts, watcher.Command{Type: "before", Command: val})
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Launch from a specific path",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Path = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[y/n]"),
													Msg:     "Tag as global command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Global = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[y/n]"),
													Msg:     "Display command output",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Output = val
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											if val {
												d.Reload()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(none)"))
											d.SetEnd("!")
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
											Msg:     "Add an 'after' custom commands  (insert '!' to stop)",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Insert a command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts = append(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts, watcher.Command{Type: "after", Command: val})
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef("", style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[string]"),
													Msg:     "Launch from a specific path",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Path = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[y/n]"),
													Msg:     "Tag as global command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Global = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, style.Green.Regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: style.Yellow.Regular("[y/n]"),
													Msg:     "Display command output",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts[len(r.Blueprint.Projects[len(r.Blueprint.Projects)-1].Watcher.Scripts)-1].Output = val
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											if val {
												d.Reload()
											}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, style.Green.Regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[y/n]"),
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
											d.SetDef("", style.Green.Regular("(none)"))
											return nil
										},
										Quest: interact.Quest{
											Options: style.Yellow.Regular("[string]"),
											Msg:     "Set an error output pattern",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Blueprint.Projects[len(r.Blueprint.Projects)-1].ErrorOutputPattern = val
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
								actErr = r.Settings.Remove(settings.Directory)
								if actErr != nil {
									return actErr
								}
							}
							return nil
						},
					})
					if err := r.Record(r); err != nil {
						return err
					}
					fmt.Fprintln(style.Output, prefix(style.Green.Bold("Your configuration was successful.")))
					return nil
				},
				Before: before,
			},
			{
				Name:        "remove",
				Category:    "Configuration",
				Aliases:     []string{"r"},
				Description: "Remove a project from a realize configuration.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
				},
				Action: func(p *cli.Context) error {
					if err := r.Blueprint.Remove(p); err != nil {
						return err
					}
					if err := r.Record(r); err != nil {
						return err
					}
					fmt.Fprintln(style.Output, prefix(style.Green.Bold("Your project was successfully removed.")))
					return nil
				},
				Before: before,
			},
			{
				Name:        "list",
				Category:    "Configuration",
				Aliases:     []string{"l"},
				Description: "Print projects list.",
				Action: func(p *cli.Context) error {
					return r.Blueprint.List()
				},
				Before: before,
			},
			{
				Name:        "clean",
				Category:    "Configuration",
				Aliases:     []string{"c"},
				Description: "Remove realize folder.",
				Action: func(p *cli.Context) error {
					if err := r.Settings.Remove(settings.Directory); err != nil {
						return err
					}
					fmt.Fprintln(style.Output, prefix(style.Green.Bold("Realize folder successfully removed.")))
					return nil
				},
				Before: before,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		print(style.Red.Bold(err))
		os.Exit(1)
	}
}

// Prefix a given string
func prefix(s string) string {
	if s != "" {
		return fmt.Sprint(style.Yellow.Bold("[")+"REALIZE"+style.Yellow.Bold("]"), s)
	}
	return ""
}

// Before is launched before each command
func before(*cli.Context) error {
	// Before of every exec of a cli method
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("$GOPATH isn't set properly")
	}
	r = realize{
		Sync: make(chan string),
		Settings: settings.Settings{
			File: settings.File,
			Server: settings.Server{
				Status: false,
				Open:   false,
				Host:   server.Host,
				Port:   server.Port,
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
	if r.FileLimit != 0 {
		if err := r.Flimit(); err != nil {
			return err
		}
	}
	return nil
}

// Check for polling option
func polling(c *cli.Context, s *settings.Legacy) {
	if c.Bool("legacy") {
		s.Interval = settings.Interval
	}
}

// Insert a project if there isn't already one
func insert(c *cli.Context, b *watcher.Blueprint) error {
	if len(b.Projects) <= 0 {
		if err := b.Add(c); err != nil {
			return err
		}
	}
	return nil
}
