package realize

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

// Config struct contains the general informations about a project
type Config struct {
	file     string
	Version  string `yaml:"version,omitempty"`
	Projects []Project
}

// nameParam check the project name presence. If empty takes the working directory name
func nameParam(params *cli.Context) string{
	var name string
	if params.String("name") == "" {
		name = params.String("base")
	}else{
		name = params.String("name")
	}
	return name
}

func boolParam(b bool) bool{
	if b{
		return false
	}
	return true
}

// New method puts the cli params in the struct
func New(params *cli.Context) *Config {
	return &Config{
		file:    AppFile,
		Version: AppVersion,
		Projects: []Project{
			{
				Name:  nameParam(params),
				Path:  params.String("base"),
				Build: params.Bool("build"),
				Bin:   boolParam(params.Bool("bin")),
				Run:   boolParam(params.Bool("run")),
				Watcher: Watcher{
					Paths:  watcherPaths,
					Ignore: watcherIgnores,
					Exts:   watcherExts,
				},
			},
		},
	}
}

// Duplicates check projects with same name or same combinations of main/path
func Duplicates(value Project, arr []Project) error {
	for _, val := range arr {
		if value.Path == val.Path || value.Name == val.Name {
			return errors.New("There is a duplicate of '"+val.Name+"'. Check your config file!")
		}
	}
	return nil
}

// Clean duplicate projects
func (h *Config) Clean() {
	arr := h.Projects
	for key, val := range arr {
		if err := Duplicates(val, arr[key+1:]); err != nil {
			h.Projects = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Read, Check and remove duplicates from the config file
func (h *Config) Read() error {
	_, err := os.Stat(h.file)
	if err == nil {
		file, err := ioutil.ReadFile(h.file)
		if err == nil {
			if len(h.Projects) > 0 {
				err = yaml.Unmarshal(file, h)
				if err == nil {
					h.Clean()
				}
				return err
			}
			return errors.New("There are no projects")
		}
		return h.Create()
	}
	return err
}

// Create and unmarshal yaml config file
func (h *Config) Create() error {
	y, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(h.file, y, 0755)
}

// Add another project
func (h *Config) Add(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		new := Project{
			Name:  nameParam(params),
			Path:  params.String("base"),
			Build: params.Bool("build"),
			Bin:   boolParam(params.Bool("bin")),
			Run:   boolParam(params.Bool("run")),
			Watcher: Watcher{
				Paths:  watcherPaths,
				Exts:   watcherExts,
				Ignore: watcherIgnores,
			},
		}
		if err := Duplicates(new, h.Projects); err != nil {
			return err
		}
		h.Projects = append(h.Projects, new)
		err = h.Create()
		if err == nil {
			fmt.Println(Green("Your project was successfully added"))
		}
		return err
	}
	err = h.Create()
	if err == nil{
		fmt.Println(Green("The config file was successfully created"))
	}
	return err
}

// Remove a project in list
func (h *Config) Remove(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		for key, val := range h.Projects {
			if params.String("name") == val.Name {
				h.Projects = append(h.Projects[:key], h.Projects[key+1:]...)
				err = h.Create()
				if err == nil {
					fmt.Println(Green("Your project was successfully removed"))
				}
				return err
			}
		}
		return errors.New("No project found")
	}
	return err
}

// List of projects
func (h *Config) List() error {
	err := h.Read()
	if err == nil {
		for _, val := range h.Projects {
			fmt.Println(Blue("|"), Blue(strings.ToUpper(val.Name)))
			fmt.Println(MagentaS("|"), "\t", Yellow("Base Path"), ":", MagentaS(val.Path))
			fmt.Println(MagentaS("|"), "\t", Yellow("Run"), ":", MagentaS(val.Run))
			fmt.Println(MagentaS("|"), "\t", Yellow("Build"),":", MagentaS(val.Build))
			fmt.Println(MagentaS("|"), "\t", Yellow("Install"), ":", MagentaS(val.Bin))
			fmt.Println(MagentaS("|"), "\t", Yellow("Watcher"),":")
			fmt.Println(MagentaS("|"), "\t\t", Yellow("After"), ":", MagentaS(val.Watcher.After))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Before"), ":", MagentaS(val.Watcher.Before))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Extensions"), ":", MagentaS(val.Watcher.Exts))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths"), ":", MagentaS(val.Watcher.Paths))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Paths ignored"), ":", MagentaS(val.Watcher.Ignore))
			fmt.Println(MagentaS("|"), "\t\t", Yellow("Watch preview"), ":", MagentaS(val.Watcher.Preview))
		}
		return nil
	}
	return err
}
