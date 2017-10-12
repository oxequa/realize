package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/tockins/interact"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	version = "1.5.0"
)

// New realize instance
var r realize

// Log struct
type logWriter struct{}

// Realize struct contains the general app informations
type realize struct {
	Settings Settings  `yaml:"settings"`
	Server   Server    `yaml:"server"`
	Schema   []Project `yaml:"schema"`
	sync     chan string
}

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
		Description: "Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Commands: []*cli.Command{
			{
				Name:        "start",
				Aliases:     []string{"r"},
				Description: "Start a toolchain on a project or a list of projects. If not exist a config file it creates a new one",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path"},
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
				Action: func(p *cli.Context) error {
					if err := r.insert(p); err != nil {
						return err
					}
					if !p.Bool("no-config") {
						if err := r.Settings.record(r); err != nil {
							return err
						}
					}
					if err := r.Server.start(p); err != nil {
						return err
					}
					if err := r.run(p); err != nil {
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
				Description: "Add a project to an existing config file or create a new one",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: "", Usage: "Project base path"},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "vet", Aliases: []string{"v"}, Value: false, Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
				},
				Action: func(p *cli.Context) error {
					if err := r.add(p); err != nil {
						return err
					}
					if err := r.Settings.record(r); err != nil {
						return err
					}
					fmt.Fprintln(output, prefix(green.bold("Your project was successfully added")))
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
							context.SetErr(red.bold("INVALID INPUT"))
							context.SetPrfx(color.Output, yellow.bold("[")+"REALIZE"+yellow.bold("]"))
							return nil
						},
						Questions: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									if _, err := os.Stat(directory + "/" + file); err != nil {
										d.Skip()
									}
									d.SetDef(false, green.regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: yellow.regular("[y/n]"),
									Msg:     "Would you want to overwrite existing " + magenta.bold("Realize") + " config?",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									} else if val {
										r = new()
									}
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, green.regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: yellow.regular("[y/n]"),
									Msg:     "Would you want to customize settings?",
									Resolve: func(d interact.Context) bool {
										val, _ := d.Ans().Bool()
										return val
									},
								},
								Subs: []*interact.Question{
									{
										Before: func(d interact.Context) error {
											d.SetDef(0, green.regular("(os default)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[int]"),
											Msg:     "Set max number of open files (root required)",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Int()
											if err != nil {
												return d.Err()
											}
											r.Settings.FileLimit = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Force polling watcher?",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(100, green.regular("(100ms)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[int]"),
													Msg:     "Set polling interval",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Int()
													if err != nil {
														return d.Err()
													}
													r.Settings.Legacy.Interval = time.Duration(int(val)) * time.Millisecond
													return nil
												},
											},
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Settings.Legacy.Force = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable logging files",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Settings.Files.Errors = Resource{Name: fileErr, Status: val}
											r.Settings.Files.Outputs = Resource{Name: fileOut, Status: val}
											r.Settings.Files.Logs = Resource{Name: fileLog, Status: val}
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable web server",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef(port, green.regular("("+strconv.Itoa(port)+")"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[int]"),
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
													d.SetDef(host, green.regular("("+host+")"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
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
													d.SetDef(false, green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[y/n]"),
													Msg:     "Open in current browser",
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
									d.SetDef(true, green.regular("(y)"))
									d.SetEnd("!")
									return nil
								},
								Quest: interact.Quest{
									Options: yellow.regular("[y/n]"),
									Msg:     "Would you want to " + magenta.regular("add a new project") + "? (insert '!' to stop)",
									Resolve: func(d interact.Context) bool {
										val, _ := d.Ans().Bool()
										if val {
											r.add(p)
										}
										return val
									},
								},
								Subs: []*interact.Question{
									{
										Before: func(d interact.Context) error {
											d.SetDef(r.Settings.wdir(), green.regular("("+r.Settings.wdir()+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[string]"),
											Msg:     "Project name",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Schema[len(r.Schema)-1].Name = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											dir, _ := os.Getwd()
											d.SetDef(dir, green.regular("("+dir+")"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[string]"),
											Msg:     "Project path",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Schema[len(r.Schema)-1].Path = r.Settings.path(val)
											return nil
										},
									},

									{
										Before: func(d interact.Context) error {
											d.SetDef(true, green.regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go vet",
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Vet additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Vet.Args = append(r.Schema[len(r.Schema)-1].Cmds.Vet.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Vet.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, green.regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go fmt",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Fmt additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Fmt.Args = append(r.Schema[len(r.Schema)-1].Cmds.Fmt.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Fmt.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, green.regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go test",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Test additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Test.Args = append(r.Schema[len(r.Schema)-1].Cmds.Test.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Test.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go generate",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Generate additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Generate.Args = append(r.Schema[len(r.Schema)-1].Cmds.Generate.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Generate.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, green.regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go install",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Install additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Install.Args = append(r.Schema[len(r.Schema)-1].Cmds.Install.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Install.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go build",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												return val
											},
										},
										Subs: []*interact.Question{
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(none)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Build additional arguments",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													if val != "" {
														r.Schema[len(r.Schema)-1].Cmds.Build.Args = append(r.Schema[len(r.Schema)-1].Cmds.Build.Args, val)
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
											r.Schema[len(r.Schema)-1].Cmds.Build.Status = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(true, green.regular("(y)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Enable go run",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Schema[len(r.Schema)-1].Cmds.Run = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Customize watching paths",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												if val {
													r.Schema[len(r.Schema)-1].Watcher.Paths = r.Schema[len(r.Schema)-1].Watcher.Paths[:len(r.Schema[len(r.Schema)-1].Watcher.Paths)-1]
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
													Options: yellow.regular("[string]"),
													Msg:     "Insert a path to watch (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Paths = append(r.Schema[len(r.Schema)-1].Watcher.Paths, val)
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
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Customize ignore paths",
											Resolve: func(d interact.Context) bool {
												val, _ := d.Ans().Bool()
												if val {
													r.Schema[len(r.Schema)-1].Watcher.Ignore = r.Schema[len(r.Schema)-1].Watcher.Ignore[:len(r.Schema[len(r.Schema)-1].Watcher.Ignore)-1]
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
													Options: yellow.regular("[string]"),
													Msg:     "Insert a path to ignore (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Ignore = append(r.Schema[len(r.Schema)-1].Watcher.Ignore, val)
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
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
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
													Options: yellow.regular("[string]"),
													Msg:     "Add another argument (insert '!' to stop)",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Args = append(r.Schema[len(r.Schema)-1].Args, val)
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
											d.SetDef(false, green.regular("(none)"))
											d.SetEnd("!")
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
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
													Options: yellow.regular("[string]"),
													Msg:     "Insert a command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts = append(r.Schema[len(r.Schema)-1].Watcher.Scripts, Command{Type: "before", Command: val})
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Launch from a specific path",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Path = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[y/n]"),
													Msg:     "Tag as global command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Global = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[y/n]"),
													Msg:     "Display command output",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Output = val
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
											d.SetDef(false, green.regular("(none)"))
											d.SetEnd("!")
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
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
													Options: yellow.regular("[string]"),
													Msg:     "Insert a command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts = append(r.Schema[len(r.Schema)-1].Watcher.Scripts, Command{Type: "after", Command: val})
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef("", green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[string]"),
													Msg:     "Launch from a specific path",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().String()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Path = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[y/n]"),
													Msg:     "Tag as global command",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Global = val
													return nil
												},
											},
											{
												Before: func(d interact.Context) error {
													d.SetDef(false, green.regular("(n)"))
													return nil
												},
												Quest: interact.Quest{
													Options: yellow.regular("[y/n]"),
													Msg:     "Display command output",
												},
												Action: func(d interact.Context) interface{} {
													val, err := d.Ans().Bool()
													if err != nil {
														return d.Err()
													}
													r.Schema[len(r.Schema)-1].Watcher.Scripts[len(r.Schema[len(r.Schema)-1].Watcher.Scripts)-1].Output = val
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
											d.SetDef(false, green.regular("(n)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[y/n]"),
											Msg:     "Print watched files on startup",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().Bool()
											if err != nil {
												return d.Err()
											}
											r.Schema[len(r.Schema)-1].Watcher.Preview = val
											return nil
										},
									},
									{
										Before: func(d interact.Context) error {
											d.SetDef("", green.regular("(none)"))
											return nil
										},
										Quest: interact.Quest{
											Options: yellow.regular("[string]"),
											Msg:     "Set an error output pattern",
										},
										Action: func(d interact.Context) interface{} {
											val, err := d.Ans().String()
											if err != nil {
												return d.Err()
											}
											r.Schema[len(r.Schema)-1].ErrorOutputPattern = val
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
								actErr = r.Settings.del(directory)
								if actErr != nil {
									return actErr
								}
							}
							return nil
						},
					})
					if err := r.Settings.record(r); err != nil {
						return err
					}
					fmt.Fprintln(output, prefix(green.bold(" Your configuration was successful")))
					return nil
				},
				Before: before,
			},
			{
				Name:        "remove",
				Category:    "Configuration",
				Aliases:     []string{"r"},
				Description: "Remove a project from a realize configuration",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: ""},
				},
				Action: func(p *cli.Context) error {
					if err := r.remove(p); err != nil {
						return err
					}
					if err := r.Settings.record(r); err != nil {
						return err
					}
					fmt.Fprintln(output, prefix(green.bold("Your project was successfully removed")))
					return nil
				},
				Before: before,
			},
			{
				Name:        "clean",
				Category:    "Configuration",
				Aliases:     []string{"c"},
				Description: "Remove realize folder",
				Action: func(p *cli.Context) error {
					if err := r.Settings.del(directory); err != nil {
						return err
					}
					fmt.Fprintln(output, prefix(green.bold("Realize folder successfully removed")))
					return nil
				},
				Before: before,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(output, prefix(red.bold(err)))
		os.Exit(1)
	}
}

// New return default realize config
func new() realize {
	return realize{
		sync: make(chan string),
		Settings: Settings{
			file: file,
			Legacy: Legacy{
				Interval: 100 * time.Millisecond,
			},
		},
		Server: Server{
			parent: &r,
			Status: false,
			Open:   false,
			Host:   host,
			Port:   port,
		},
	}
}

// Prefix a given string
func prefix(s string) string {
	if s != "" {
		return fmt.Sprint(yellow.bold("["), "REALIZE", yellow.bold("]"), s)
	}
	return ""
}

// Before is launched before each command
func before(*cli.Context) error {
	// custom log
	log.SetFlags(0)
	log.SetOutput(logWriter{})
	// Before of every exec of a cli method
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("$GOPATH isn't set properly")
	}
	// new realize instance
	r = new()
	// read if exist
	r.Settings.read(&r)
	// increase the file limit
	if r.Settings.FileLimit != 0 {
		if err := r.Settings.flimit(); err != nil {
			return err
		}
	}
	return nil
}

// Rewrite the layout of the log timestamp
func (w logWriter) Write(bytes []byte) (int, error) {
	return fmt.Fprint(output, yellow.regular("["), time.Now().Format("15:04:05"), yellow.regular("]")+string(bytes))
}
