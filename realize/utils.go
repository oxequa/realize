package realize

import (
	"errors"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"strings"
)

// Params parse one by one the given argumentes
func params(params *cli.Context) []string {
	argsN := params.NArg()
	if argsN > 0 {
		var args []string
		for i := 0; i <= argsN-1; i++ {
			args = append(args, params.Args().Get(i))
		}
		return args
	}
	return nil
}

// Split each arguments in multiple fields
func split(args, fields []string) []string {
	for _, arg := range fields {
		arr := strings.Fields(arg)
		args = append(args, arr...)
	}
	return args
}

// Duplicates check projects with same name or same combinations of main/path
func duplicates(value Project, arr []Project) (Project, error) {
	for _, val := range arr {
		if value.Name == val.Name {
			return val, errors.New("There is already a project with name '" + val.Name + "'. Check your config file!")
		}
	}
	return Project{}, nil
}

// Get file extensions
func ext(path string) string {
	var ext string
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			ext = path[i:]
			if index := strings.LastIndex(ext, "."); index > 0 {
				ext = ext[index:]
			}
		}
	}
	if ext != "" {
		return ext[1:]
	}
	return ""
}

// Replace if isn't empty and create a new array
func replace(a []string, b string) []string {
	if len(b) > 0 {
		return strings.Fields(b)
	}
	return a
}

// Wdir return current working directory
func Wdir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	return dir
}
