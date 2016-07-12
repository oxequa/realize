package realize

import (
	"os"
)

type Config struct {
	App_file string
	app_main []string
	app_version string
	app_build bool
	app_run struct {
		before, after, paths, ext []string
	}
}

// Create config yaml file
func (h *Config) Create() bool{
	var config = Check(h.App_file)
	if config[0] == false {
		if _, err := os.Create("realize.config.yml"); err == nil {
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


