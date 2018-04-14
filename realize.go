package main

import (
	"github.com/oxequa/interact"
	"github.com/oxequa/realize/realize"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var r realize.Realize

// Realize cli commands
func main() {
	r.Sync = make(chan string)
	app := &cli.App{
		Name:        strings.Title(realize.RPrefix),
		Version:     realize.RVersion,
		Description: "Realize is the #1 Golang Task Runner which enhance your workflow by automating the most common tasks and using the best performing Golang live reloading.",
		Commands: []*cli.Command{
			{
				Name:        "start",
				Aliases:     []string{"s"},
				Description: "Start " + strings.Title(realize.RPrefix) + " on a given path. If not exist a config file it creates a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: ".", Usage: "Project base path"},
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "", Usage: "Run a project by its name"},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "vet", Aliases: []string{"v"}, Value: false, Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "server", Aliases: []string{"srv"}, Value: false, Usage: "Start server"},
					&cli.BoolFlag{Name: "open", Aliases: []string{"op"}, Value: false, Usage: "Open into the default browser"},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
					&cli.BoolFlag{Name: "legacy", Aliases: []string{"l"}, Value: false, Usage: "Legacy watch by polling instead fsnotify"},
					&cli.BoolFlag{Name: "no-config", Aliases: []string{"nc"}, Value: false, Usage: "Ignore existing config and doesn't create a new one"},
				},
				Action: func(c *cli.Context) error {
					return start(c)
				},
			},
			{
				Name:        "add",
				Category:    "Configuration",
				Aliases:     []string{"a"},
				Description: "Add a project to an existing config or to a new one.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Value: realize.Wdir(), Usage: "Project base path"},
					&cli.BoolFlag{Name: "fmt", Aliases: []string{"f"}, Value: false, Usage: "Enable go fmt"},
					&cli.BoolFlag{Name: "vet", Aliases: []string{"v"}, Value: false, Usage: "Enable go vet"},
					&cli.BoolFlag{Name: "test", Aliases: []string{"t"}, Value: false, Usage: "Enable go test"},
					&cli.BoolFlag{Name: "generate", Aliases: []string{"g"}, Value: false, Usage: "Enable go generate"},
					&cli.BoolFlag{Name: "install", Aliases: []string{"i"}, Value: false, Usage: "Enable go install"},
					&cli.BoolFlag{Name: "build", Aliases: []string{"b"}, Value: false, Usage: "Enable go build"},
					&cli.BoolFlag{Name: "run", Aliases: []string{"nr"}, Value: false, Usage: "Enable go run"},
				},
				Action: func(c *cli.Context) error {
					return add(c)
				},
			},
			{
				Name:        "init",
				Category:    "Configuration",
				Aliases:     []string{"i"},
				Description: "Make a new config file step by step.",
				Action: func(c *cli.Context) error {
					return setup(c)
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
					return remove(c)
				},
			},
			{
				Name:        "clean",
				Category:    "Configuration",
				Aliases:     []string{"c"},
				Description: "Remove " + strings.Title(realize.RPrefix) + " folder.",
				Action: func(c *cli.Context) error {
					return clean()
				},
			},
			{
				Name:        "version",
				Aliases:     []string{"v"},
				Description: "Print " + strings.Title(realize.RPrefix) + " version.",
				Action: func(p *cli.Context) error {
					version()
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

// Version print current version
func version() {
	log.Println(r.Prefix(realize.Green.Bold(realize.RVersion)))
}

// Clean remove realize file
func clean() (err error) {
	if err := r.Settings.Remove(realize.RFile); err != nil {
		return err
	}
	log.Println(r.Prefix(realize.Green.Bold("folder successfully removed")))
	return nil
}

// Add a project to an existing config or create a new one
func add(c *cli.Context) (err error) {
	// read a config if exist
	err = r.Settings.Read(&r)
	if err != nil {
		return err
	}
	projects := len(r.Schema.Projects)
	// create and add a new project
	r.Schema.Add(r.Schema.New(c))
	if len(r.Schema.Projects) > projects {
		// update config
		err = r.Settings.Write(r)
		if err != nil {
			return err
		}
		log.Println(r.Prefix(realize.Green.Bold("project successfully added")))
	} else {
		log.Println(r.Prefix(realize.Green.Bold("project can't be added")))
	}
	return nil
}

// Setup a new config step by step
func setup(c *cli.Context) (err error) {
	interact.Run(&interact.Interact{
		Before: func(context interact.Context) error {
			context.SetErr(realize.Red.Bold("INVALID INPUT"))
			context.SetPrfx(realize.Output, realize.Yellow.Regular("[")+time.Now().Format("15:04:05")+realize.Yellow.Regular("]")+realize.Yellow.Bold("[")+strings.ToUpper(realize.RPrefix)+realize.Yellow.Bold("]"))
			return nil
		},
		Questions: []*interact.Question{
			{
				Before: func(d interact.Context) error {
					if _, err := os.Stat(realize.RFile); err != nil {
						d.Skip()
					}
					d.SetDef(false, realize.Green.Regular("(n)"))
					return nil
				},
				Quest: interact.Quest{
					Options: realize.Yellow.Regular("[y/n]"),
					Msg:     "Would you want to overwrite existing " + realize.Magenta.Regular(realize.RPrefix) + " config?",
				},
				Action: func(d interact.Context) interface{} {
					val, err := d.Ans().Bool()
					if err != nil {
						return d.Err()
					} else if val {
						r := realize.Realize{}
						r.Server = realize.Server{Parent: &r, Status: false, Open: false, Port: realize.Port, Host: realize.Host}
					}
					return nil
				},
			},
			{
				Before: func(d interact.Context) error {
					d.SetDef(false, realize.Green.Regular("(n)"))
					return nil
				},
				Quest: interact.Quest{
					Options: realize.Yellow.Regular("[y/n]"),
					Msg:     "Would you want to customize settings?",
					Resolve: func(d interact.Context) bool {
						val, _ := d.Ans().Bool()
						return val
					},
				},
				Subs: []*interact.Question{
					{
						Before: func(d interact.Context) error {
							d.SetDef(0, realize.Green.Regular("(os default)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[int]"),
							Msg:     "Set max number of open files (root required)",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().Int()
							if err != nil {
								return d.Err()
							}
							r.Settings.FileLimit = int32(val)
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Force polling watcher?",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef(100, realize.Green.Regular("(100ms)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[int]"),
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
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable logging files",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().Bool()
							if err != nil {
								return d.Err()
							}
							r.Settings.Files.Errors = realize.Resource{Name: realize.FileErr, Status: val}
							r.Settings.Files.Outputs = realize.Resource{Name: realize.FileOut, Status: val}
							r.Settings.Files.Logs = realize.Resource{Name: realize.FileLog, Status: val}
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable web server",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef(realize.Port, realize.Green.Regular("("+strconv.Itoa(realize.Port)+")"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[int]"),
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
									d.SetDef(realize.Host, realize.Green.Regular("("+realize.Host+")"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
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
									d.SetDef(false, realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[y/n]"),
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
					d.SetDef(true, realize.Green.Regular("(y)"))
					d.SetEnd("!")
					return nil
				},
				Quest: interact.Quest{
					Options: realize.Yellow.Regular("[y/n]"),
					Msg:     "Would you want to " + realize.Magenta.Regular("add a new project") + "? (insert '!' to stop)",
					Resolve: func(d interact.Context) bool {
						val, _ := d.Ans().Bool()
						if val {
							r.Schema.Add(r.Schema.New(c))
						}
						return val
					},
				},
				Subs: []*interact.Question{
					{
						Before: func(d interact.Context) error {
							d.SetDef(realize.Wdir(), realize.Green.Regular("("+realize.Wdir()+")"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[string]"),
							Msg:     "Project name",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().String()
							if err != nil {
								return d.Err()
							}
							r.Schema.Projects[len(r.Schema.Projects)-1].Name = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							dir := realize.Wdir()
							d.SetDef(dir, realize.Green.Regular("("+dir+")"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[string]"),
							Msg:     "Project path",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().String()
							if err != nil {
								return d.Err()
							}
							r.Schema.Projects[len(r.Schema.Projects)-1].Path = filepath.Clean(val)
							return nil
						},
					},

					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go vet",
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Vet additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Vet.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Vet.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Vet.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go fmt",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Fmt additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fmt.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fmt.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fmt.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go test",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Test additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Test.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Test.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Test.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go clean",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Clean additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Clean.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Clean.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Clean.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go generate",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Generate additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Generate.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Generate.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Generate.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(true, realize.Green.Regular("(y)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go install",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Install additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Install.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Install.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Install.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go build",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								return val
							},
						},
						Subs: []*interact.Question{
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(none)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Build additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Build.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Build.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Build.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(true, realize.Green.Regular("(y)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Enable go run",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().Bool()
							if err != nil {
								return d.Err()
							}
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Run.Status = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Customize watching paths",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								if val {
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Paths = r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Paths[:len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Paths)-1]
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
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Insert a path to watch (insert '!' to stop)",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Paths = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Paths, val)
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
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
							Msg:     "Customize ignore paths",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								if val {
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore.Paths = r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore.Paths[:len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore.Paths)-1]
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
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Insert a path to ignore (insert '!' to stop)",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore.Paths = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore.Paths, val)
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
							d.SetDef(false, realize.Green.Regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
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
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Add another argument (insert '!' to stop)",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Args, val)
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
							d.SetDef(false, realize.Green.Regular("(none)"))
							d.SetEnd("!")
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
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
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Insert a command",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts, realize.Command{Type: "before", Cmd: val})
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Launch from a specific path",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Path = val
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[y/n]"),
									Msg:     "Tag as global command",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Global = val
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[y/n]"),
									Msg:     "Display command output",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Output = val
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
							d.SetDef(false, realize.Green.Regular("(none)"))
							d.SetEnd("!")
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[y/n]"),
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
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Insert a command",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts, realize.Command{Type: "after", Cmd: val})
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef("", realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[string]"),
									Msg:     "Launch from a specific path",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Path = val
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[y/n]"),
									Msg:     "Tag as global command",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Global = val
									return nil
								},
							},
							{
								Before: func(d interact.Context) error {
									d.SetDef(false, realize.Green.Regular("(n)"))
									return nil
								},
								Quest: interact.Quest{
									Options: realize.Yellow.Regular("[y/n]"),
									Msg:     "Display command output",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().Bool()
									if err != nil {
										return d.Err()
									}
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Output = val
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
							d.SetDef("", realize.Green.Regular("(none)"))
							return nil
						},
						Quest: interact.Quest{
							Options: realize.Yellow.Regular("[string]"),
							Msg:     "Set an error output pattern",
						},
						Action: func(d interact.Context) interface{} {
							val, err := d.Ans().String()
							if err != nil {
								return d.Err()
							}
							r.Schema.Projects[len(r.Schema.Projects)-1].ErrPattern = val
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
				err := r.Settings.Remove(realize.RFile)
				if err != nil {
					return err
				}
			}
			return nil
		},
	})
	// create config
	err = r.Settings.Write(r)
	if err != nil {
		return err
	}
	log.Println(r.Prefix(realize.Green.Bold("Config successfully created")))
	return nil
}

// Start realize workflow
func start(c *cli.Context) (err error) {
	// set legacy watcher
	if c.Bool("legacy") {
		r.Settings.Legacy.Set(c.Bool("legacy"), 1)
	}
	// set server
	if c.Bool("server") {
		r.Server.Set(c.Bool("server"), c.Bool("open"), realize.Port, realize.Host)
	}
	// check no-config and read
	if !c.Bool("no-config") {
		// read a config if exist
		r.Settings.Read(&r)
		if c.String("name") != "" {
			// filter by name flag if exist
			r.Schema.Projects = r.Schema.Filter("Name", c.String("name"))
		}
		// increase file limit
		if r.Settings.FileLimit != 0 {
			if err = r.Settings.Flimit(); err != nil {
				return err
			}
		}

	}
	// check project list length
	if len(r.Schema.Projects) <= 0 {
		println("len", r.Schema.Projects)
		// create a new project based on given params
		project := r.Schema.New(c)
		// Add to projects list
		r.Schema.Add(project)
		// save config
		if !c.Bool("no-config") {
			err = r.Settings.Write(r)
			if err != nil {
				return err
			}
		}
	}
	// Start web server
	if r.Server.Status {
		r.Server.Parent = &r
		err = r.Server.Start()
		if err != nil {
			return err
		}
		err = r.Server.OpenURL()
		if err != nil {
			return err
		}
	}
	// start workflow
	return r.Start()
}

// Remove a project from an existing config
func remove(c *cli.Context) (err error) {
	// read a config if exist
	err = r.Settings.Read(&r)
	if err != nil {
		return err
	}
	if c.String("name") != "" {
		err := r.Schema.Remove(c.String("name"))
		if err != nil {
			return err
		}
		// update config
		err = r.Settings.Write(r)
		if err != nil {
			return err
		}
		log.Println(r.Prefix(realize.Green.Bold("project successfully removed")))
	} else {
		log.Println(r.Prefix(realize.Green.Bold("project name not found")))
	}
	return nil
}
