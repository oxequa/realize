package realize

import (
	"os"
	"gopkg.in/yaml.v2"
	"errors"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
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
	Watcher Watcher `yaml:"app_watcher,omitempty"`
}

type Watcher struct{
	Before []string `yaml:"before,omitempty"`
	After []string `yaml:"after,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
	Exts []string `yaml:"exts,omitempty"`
}

var file = "realize.config.yaml"

// Check files exists
func Check(files ...string) (result []bool){
	for _, val := range files {
		if _, err := os.Stat(val); err == nil {
			result = append(result,true)
		}
		result = append(result, false)
	}
	return
}

// Default value
func (h *Config) Init(params *cli.Context) {
	h.file = file
	h.Version = "1.0"
	h.Projects = []Project{
		{
			Main: params.String("main"),
			Run: params.Bool("run"),
			Build: params.Bool("build"),
			Watcher: Watcher{
				Paths: []string{"/"},
				Exts: []string{"go"},
			},
		},
	}
}

// Read config file
func (h *Config) Read() error{
	if filename, err := filepath.Abs("./"+file); err == nil{
		y, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(y, &h)
	}else{
		return err
	}
}

// Create config yaml file
func (h *Config) Create() error{
	config := Check(h.file)
	if config[0] == false {
		if w, err := os.Create(h.file); err == nil {
			y, err := yaml.Marshal(h)
			if err != nil {
				os.Remove(h.file)
				return err
			}
			_, err = w.WriteString(string(y))
			return err
		}
		return errors.New("There is a problem with the file's creation")
	}
	return errors.New("The configuration file already exist")
}

// Add another project
func (h *Config) Add(params *cli.Context) {
	new := Project{
		Main: params.String("main"),
		Run: params.Bool("run"),
		Build: params.Bool("build"),
		Watcher: Watcher{
			Paths: []string{"/"},
			Exts: []string{"go"},
		},
	}
	h.Projects = append(h.Projects, new)
}

