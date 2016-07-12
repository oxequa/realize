package realize

import (
	"os"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App_file string
	App_main []string
	App_version string
	App_build bool
	App_run struct {
		before, after, paths, ext []string
	}
}

// Create config yaml file
func (h *Config) Create() bool{
	var config = Check(h.App_file)
	if config[0] == false {
		if w, err := os.Create(h.App_file); err == nil {
			defer w.Close()
			y, err := yaml.Marshal(&h)
			if err != nil {
				defer panic(err)
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


