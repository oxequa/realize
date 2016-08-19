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
				Main:  params.String("main"),
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
		if value.Main == val.Main && value.Path == val.Path || value.Name == val.Name {
			Fail("There is a duplicate of '"+val.Name+"'. Check your config file!")
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
			Success("The config file was successfully created")
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
			Main:  params.String("main"),
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
			Success("Your project was successfully added")
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
					Success("Your project was successfully removed")
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
			fmt.Println(green("|"), green(val.Name))
			fmt.Println(greenl("|"), "\t", green("Main File:"), red(val.Main))
			fmt.Println(greenl("|"), "\t", green("Base Path:"), red(val.Path))
			fmt.Println(greenl("|"), "\t", green("Run:"), red(val.Run))
			fmt.Println(greenl("|"), "\t", green("Build:"), red(val.Build))
			fmt.Println(greenl("|"), "\t", green("Install:"), red(val.Bin))
			fmt.Println(greenl("|"), "\t", green("Watcher:"))
			fmt.Println(greenl("|"), "\t\t", green("After:"), red(val.Watcher.After))
			fmt.Println(greenl("|"), "\t\t", green("Before:"), red(val.Watcher.Before))
			fmt.Println(greenl("|"), "\t\t", green("Extensions:"), red(val.Watcher.Exts))
			fmt.Println(greenl("|"), "\t\t", green("Paths:"), red(val.Watcher.Paths))
			fmt.Println(greenl("|"), "\t\t", green("Paths ignored:"), red(val.Watcher.Ignore))
			fmt.Println(greenl("|"), "\t\t", green("Watch preview:"), red(val.Watcher.Preview))
		}
		return nil
	}
	return err
}
