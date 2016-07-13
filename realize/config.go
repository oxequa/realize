package realize

import (
	"os"
	"gopkg.in/yaml.v2"
	"errors"
)

type Config struct {
	File string `yaml:"app_file,omitempty"`
	Main []string `yaml:"app_main,omitempty"`
	Version string `yaml:"app_version,omitempty"`
	Build bool `yaml:"app_build,omitempty"`
	Watchers
}

type Watchers struct{
	Before []string `yaml:"app_before,omitempty"`
	After []string `yaml:"app_after,omitempty"`
	Paths []string `yaml:"app_paths,omitempty"`
	Ext []string `yaml:"app_ext,omitempty"`
}

// Check files exists
func Check(files ...string) (result []bool, err error){
	for _, val := range files {
		if _, err := os.Stat(val); err == nil {
			result = append(result,true)
		}
		result = append(result, false)
	}
	return
}

// Default value
func Init() Config{
	config := Config{
		File:"realize.config.yaml",
		Main:[]string{"main.go"},
		Version:"1.0",
		Build: true,
		Watchers: Watchers{
			Paths: []string{"/"},
			Ext: []string{"go"},
		},
	}
	return config
}

// Create config yaml file
func (h *Config) Create() (result bool, err error){
	config, err := Check(h.File)
	if config[0] == false {
		if w, err := os.Create(h.File); err == nil {
			defer w.Close()
			y, err := yaml.Marshal(h)
			_, err = w.WriteString(string(y))
			if err != nil {
				os.Remove(h.File)
				return false, err
			}
			return true, nil
		}
		return false, err
	}
	return false, errors.New("already exist")
}

// Read config file
func (h *Config) Read(field string) bool {
	return true
}

