package app

import (
	c "github.com/tockins/realize/cli"
	s "github.com/tockins/realize/server"
	"syscall"
	"fmt"
)

var R Realize

// Realize struct contains the general app informations
type Realize struct {
	Name, Description, Author, Email string
	Version                          string
	Limit                            uint64
	Blueprint                        c.Blueprint
	Server				 s.Server
}

// Application initialization
func init(){
	R = Realize{
		Name:        "Realize",
		Version:     "1.0",
		Description: "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Limit:       10000,
		Blueprint: c.Blueprint{
			Files: map[string]string{
				"config": "r.config.yaml",
				"output": "r.output.log",
			},
		},
		Server:	s.Server{
			Blueprint: &R.Blueprint,
		},
	}
	R.Increases()
}

// Flimit defines the max number of watched files
func (r *Realize) Increases() {
	// increases the files limit
	var rLimit syscall.Rlimit
	rLimit.Max = r.Limit
	rLimit.Cur = r.Limit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println(c.Red("Error Setting Rlimit "), err)
	}
}
