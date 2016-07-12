package realize

import (
	"os"
	"gopkg.in/yaml.v2"
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
func Check(files ...string) []bool{
	var result []bool
	for _, val := range files {
		if _, err := os.Stat(val); err == nil {
			result = append(result,true)
		}
		result = append(result, false)
	}
	return result
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
func (h *Config) Create() bool{
	var config = Check(h.File)
	if config[0] == false {
		if w, err := os.Create(h.File); err == nil {
			defer w.Close()
			y, err := yaml.Marshal(h)
			if err != nil {
				panic(err)
			}
			w.WriteString(string(y))
			return true
		}else{
			panic(err)
		}
	}
	return false
}

// Read config file
func (h *Config) Read(field string) bool {
	return true
}

