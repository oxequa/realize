package main

import (
	"github.com/fatih/color"
	"github.com/tockins/interact"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Version print current version
func (r *Realize) version() {
	log.Println(r.Prefix(green.bold(RVersion)))
}

// Clean remove realize folder
func (r *Realize) clean() (err error) {
	if err := r.Settings.Remove(RDir); err != nil {
		return err
	}
	log.Println(r.Prefix(green.bold("folder successfully removed")))
	return nil
}

// Add a project to an existing config or create a new one
func (r *Realize) add(c *cli.Context) (err error) {
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
		log.Println(r.Prefix(green.bold("project successfully added")))
	} else {
		log.Println(r.Prefix(green.bold("project can't be added")))
	}
	return nil
}

// Setup a new config step by step
func (r *Realize) setup(c *cli.Context) (err error) {
	interact.Run(&interact.Interact{
		Before: func(context interact.Context) error {
			context.SetErr(red.bold("INVALID INPUT"))
			context.SetPrfx(color.Output, yellow.regular("[")+time.Now().Format("15:04:05")+yellow.regular("]")+yellow.bold("[")+strings.ToUpper(RPrefix)+yellow.bold("]"))
			return nil
		},
		Questions: []*interact.Question{
			{
				Before: func(d interact.Context) error {
					if _, err := os.Stat(RDir + "/" + RFile); err != nil {
						d.Skip()
					}
					d.SetDef(false, green.regular("(n)"))
					return nil
				},
				Quest: interact.Quest{
					Options: yellow.regular("[y/n]"),
					Msg:     "Would you want to overwrite existing " + magenta.bold(RPrefix) + " config?",
				},
				Action: func(d interact.Context) interface{} {
					val, err := d.Ans().Bool()
					if err != nil {
						return d.Err()
					} else if val {
						r := Realize{}
						r.Server = Server{&r, false, false, Port, Host}
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
							r.Settings.FileLimit = int32(val)
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
							r.Settings.Files.Errors = Resource{Name: FileErr, Status: val}
							r.Settings.Files.Outputs = Resource{Name: FileOut, Status: val}
							r.Settings.Files.Logs = Resource{Name: FileLog, Status: val}
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
									d.SetDef(Port, green.regular("("+strconv.Itoa(Port)+")"))
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
									d.SetDef(Host, green.regular("("+Host+")"))
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
							r.Schema.Add(r.Schema.New(c))
						}
						return val
					},
				},
				Subs: []*interact.Question{
					{
						Before: func(d interact.Context) error {
							d.SetDef(wdir(), green.regular("("+wdir()+")"))
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Name = val
							return nil
						},
					},
					{
						Before: func(d interact.Context) error {
							dir := wdir()
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Path = filepath.Clean(val)
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
							d.SetDef(false, green.regular("(n)"))
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
							d.SetDef(false, green.regular("(n)"))
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
							d.SetDef(false, green.regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: yellow.regular("[y/n]"),
							Msg:     "Enable go fix",
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
									Msg:     "Fix additional arguments",
								},
								Action: func(d interact.Context) interface{} {
									val, err := d.Ans().String()
									if err != nil {
										return d.Err()
									}
									if val != "" {
										r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fix.Args = append(r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fix.Args, val)
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Fix.Status = val
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
							Msg:     "Enable go clean",
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
							r.Schema.Projects[len(r.Schema.Projects)-1].Tools.Run = val
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
									Options: yellow.regular("[string]"),
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
							d.SetDef(false, green.regular("(n)"))
							return nil
						},
						Quest: interact.Quest{
							Options: yellow.regular("[y/n]"),
							Msg:     "Customize ignore paths",
							Resolve: func(d interact.Context) bool {
								val, _ := d.Ans().Bool()
								if val {
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore = r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore[:len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore)-1]
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Ignore, val)
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts, Command{Type: "before", Cmd: val})
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Path = val
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Global = val
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts = append(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts, Command{Type: "after", Cmd: val})
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Path = val
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
									r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts[len(r.Schema.Projects[len(r.Schema.Projects)-1].Watcher.Scripts)-1].Global = val
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
							r.Schema.Projects[len(r.Schema.Projects)-1].ErrorOutputPattern = val
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
				err := r.Settings.Remove(RDir)
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
	log.Println(r.Prefix(green.bold("Config successfully created")))
	return nil
}

// Start realize workflow
func (r *Realize) start(c *cli.Context) (err error) {
	r.Server = Server{r, false, false, Port, Host}
	// check no-config and read
	if !c.Bool("no-config") {
		// read a config if exist
		err = r.Settings.Read(&r)
		if err != nil {
			return err
		}
		if c.String("name") != "" {
			// filter by name flag if exist
			r.Schema.Filter("name", c.String("name"))
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
	// config and start server
	if c.Bool("server") || r.Server.Status {
		r.Server.Status = true
		if c.Bool("open") || r.Server.Open {
			r.Server.Open = true
			r.Server.OpenURL()
		}
		err = r.Server.Start()
		if err != nil {
			return err
		}
	}
	// start workflow
	r.Start()
	return
}

// Remove a project from an existing config
func (r *Realize) remove(c *cli.Context) (err error) {
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
		log.Println(r.Prefix(green.bold("project successfully removed")))
	} else {
		log.Println(r.Prefix(green.bold("project name not found")))
	}
	return nil
}
