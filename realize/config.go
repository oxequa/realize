package realize

import (
	"os"
	"gopkg.in/yaml.v2"
	"errors"
	"gopkg.in/urfave/cli.v2"
	"io/ioutil"
)

type Config struct {
	file string `yaml:"app_file,omitempty"`
	Version string `yaml:"version,omitempty"`
	Projects []Project
}

type Project struct {
	Run bool `yaml:"app_run,omitempty"`
	Build bool `yaml:"app_build,omitempty"`
	Main string `yaml:"app_main,omitempty"`
	Name string `yaml:"app_name,omitempty"`
	Watcher Watcher `yaml:"app_watcher,omitempty"`
}

type Watcher struct{
	Before []string `yaml:"before,omitempty"`
	After []string `yaml:"after,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
	Exts []string `yaml:"exts,omitempty"`
}

// Default value
func New(params *cli.Context) *Config{
	return &Config{
		file: "realize.config.yaml",
		Version: "1.0",
		Projects: []Project{
			{
				Main: params.String("main"),
				Run: params.Bool("run"),
				Build: params.Bool("build"),
				Watcher: Watcher{
				Paths: []string{"/"},
				Exts: []string{"go"},
				},
			},
		},
	}
}

// check for duplicates
func Duplicates(value string, arr []Project) bool{
	for _, val := range arr{
		if value == val.Main{
			return true
		}
	}
	return false
}

// Remove duplicate projects
func (h *Config) Clean() {
	arr := h.Projects
	for key, val := range arr {
		 if Duplicates(val.Main, arr[key+1:]) {
			 h.Projects = append(arr[:key], arr[key+1:]...)
			 break
		 }
	}
}

// Check, Read and remove duplicates from the config file
func (h *Config) Read() error{
	if file, err :=  ioutil.ReadFile(h.file); err == nil{
		err = yaml.Unmarshal(file, h)
		if err == nil {
			h.Clean()
		}
		return err
	}else{
		return err
	}
}

// Create config yaml file
func (h *Config) Create(params *cli.Context) error{
	if h.Read() != nil {
		if y, err := yaml.Marshal(h); err == nil {
			err = ioutil.WriteFile(h.file, y, 0755)
			if err != nil {
				os.Remove(h.file)
				return err
			}
			return err
		}else{
			return err
		}
	}
	return errors.New("The config file already exists, check for realize.config.yaml")
}

// Add another project
func (h *Config) Add(params *cli.Context) error{
	if h.Read() == nil {
		new := Project{
			Main: params.String("main"),
			Run: params.Bool("run"),
			Build: params.Bool("build"),
			Watcher: Watcher{
				Paths: []string{"/"},
				Exts: []string{"go"},
			},
		}
		if Duplicates(new.Main, h.Projects) {
			return errors.New("There is already one project with same main path")
		}
		h.Projects = append(h.Projects, new)
		y, err := yaml.Marshal(h)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(h.file, y, 0755)
	}
	return errors.New("The configuration file doesn't exist")
}


