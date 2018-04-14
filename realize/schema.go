package realize

import (
	"errors"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
	"reflect"
)

// Schema projects list
type Schema struct {
	Projects []Project `yaml:"schema" json:"schema"`
}

// Add a project if unique
func (s *Schema) Add(p Project) {
	for _, val := range s.Projects {
		if reflect.DeepEqual(val, p) {
			return
		}
	}
	s.Projects = append(s.Projects, p)
}

// Remove a project
func (s *Schema) Remove(name string) error {
	for key, val := range s.Projects {
		if name == val.Name {
			s.Projects = append(s.Projects[:key], s.Projects[key+1:]...)
			return nil
		}
	}
	return errors.New("project not found")
}

// New create a project using cli fields
func (s *Schema) New(c *cli.Context) Project {
	name := filepath.Base(c.String("path"))
	if len(name) == 0 || name == "." {
		name = filepath.Base(Wdir())
	}
	project := Project{
		Name: name,
		Path: c.String("path"),
		Tools: Tools{
			Vet: Tool{
				Status: c.Bool("vet"),
			},
			Fmt: Tool{
				Status: c.Bool("fmt"),
			},
			Test: Tool{
				Status: c.Bool("test"),
			},
			Generate: Tool{
				Status: c.Bool("generate"),
			},
			Build: Tool{
				Status: c.Bool("build"),
			},
			Install: Tool{
				Status: c.Bool("install"),
			},
			Run: Tool{
				Status: c.Bool("run"),
			},
		},
		Args: params(c),
		Watcher: Watch{
			Paths:  []string{"/"},
			Ignore: Ignore{
				Paths:[]string{".git", ".realize", "vendor"},
			},
			Exts:   []string{"go"},
		},
	}
	return project
}

// Filter project list by field
func (s *Schema) Filter(field string, value interface{}) []Project {
	result := []Project{}
	for _, item := range s.Projects {
		v := reflect.ValueOf(item)
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == field {
				if reflect.DeepEqual(v.Field(i).Interface(), value) {
					result = append(result, item)
				}
			}
		}
	}
	return result
}
