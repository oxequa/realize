package realize

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Config struct contains the general informations about a project
type Config struct {
	file     string
	Version  string `yaml:"version,omitempty"`
	Projects []Project
}

// New method puts the cli params in the struct
func New(params *cli.Context) *Config {
	return &Config{
		file:    AppFile,
		Version: AppVersion,
		Projects: []Project{
			{
				Name:  params.String("name"),
				Path:  params.String("base"),
				Run:   params.Bool("run"),
				Build: params.Bool("build"),
				Bin:   params.Bool("bin"),
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
func Duplicates(value Project, arr []Project) bool {
	for _, val := range arr {
		if value.Path == val.Path || value.Name == val.Name {
			fmt.Println(Red("There is a duplicate of '"+val.Name+"'. Check your config file!"))
			return true
		}
	}
	return false
}

// Clean duplicate projects
func (h *Config) Clean() {
	arr := h.Projects
	for key, val := range arr {
		if Duplicates(val, arr[key+1:]) {
			h.Projects = append(arr[:key], arr[key+1:]...)
			break
		}
	}
}

// Read, Check and remove duplicates from the config file
func (h *Config) Read() error {
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
	return err
}

// write and marshal yaml
func (h *Config) Write() error {
	y, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(h.file, y, 0755)
}

// Create config yaml file
func (h *Config) Create(params *cli.Context) error {
	if h.Read() != nil {
		err := h.Write()
		if err != nil {
			os.Remove(h.file)
		} else {
			fmt.Println(Green("The config file was successfully created"))
		}
		return err
	}
	return errors.New("The config file already exists, check for realize.config.yaml")
}

// Add another project
func (h *Config) Add(params *cli.Context) error {
	err := h.Read()
	if err == nil {
		new := Project{
			Name:  params.String("name"),
			Path:  params.String("base"),
			Run:   params.Bool("run"),
			Build: params.Bool("build"),
			Bin:   params.Bool("bin"),
			Watcher: Watcher{
				Paths:  watcherPaths,
				Exts:   watcherExts,
				Ignore: watcherIgnores,
			},
		}
		if Duplicates(new, h.Projects) {
			return errors.New("There is already one project with same path or name")
		}
		h.Projects = append(h.Projects, new)
		err = h.Write()
		if err == nil {
			fmt.Println(Green("Your project was successfully added"))
		}
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
				err = h.Write()
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
			fmt.Println(Green("|"), Green(val.Name))
			fmt.Println(Magenta("|"), "\t", Green("Base Path"), ":", Magenta(val.Path))
			fmt.Println(Magenta("|"), "\t", Green("Run"), ":", Magenta(val.Run))
			fmt.Println(Magenta("|"), "\t", Green("Build"),":", Magenta(val.Build))
			fmt.Println(Magenta("|"), "\t", Green("Install"), ":", Magenta(val.Bin))
			fmt.Println(Magenta("|"), "\t", Green("Watcher"),":")
			fmt.Println(Magenta("|"), "\t\t", Green("After"), ":", Magenta(val.Watcher.After))
			fmt.Println(Magenta("|"), "\t\t", Green("Before"), ":", Magenta(val.Watcher.Before))
			fmt.Println(Magenta("|"), "\t\t", Green("Extensions"), ":", Magenta(val.Watcher.Exts))
			fmt.Println(Magenta("|"), "\t\t", Green("Paths"), ":", Magenta(val.Watcher.Paths))
			fmt.Println(Magenta("|"), "\t\t", Green("Paths ignored"), ":", Magenta(val.Watcher.Ignore))
			fmt.Println(Magenta("|"), "\t\t", Green("Watch preview"), ":", Magenta(val.Watcher.Preview))
		}
		return nil
	}
	return err
}
