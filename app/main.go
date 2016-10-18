package app

import (
	"fmt"
	c "github.com/tockins/realize/cli"
	s "github.com/tockins/realize/server"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"syscall"
)

const (
	Name        = "Realize"
	Version     = "1.1"
	Description = "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths"
	Limit       = 10000
	Config      = "r.config.yaml"
	Output      = "r.output.log"
	Host        = "Web server listening on localhost:5000"
)

var r realize
var R Realizer

// Realizer interface for wrap the cli, app and server functions
type Realizer interface {
	Wdir() string
	Red(string) string
	Blue(string) string
	BlueS(string) string
	Handle(error) error
	Serve(*cli.Context)
	Before(*cli.Context) error
	Fast(*cli.Context) error
	Run(*cli.Context) error
	Add(*cli.Context) error
	Remove(*cli.Context) error
	List(*cli.Context) error
}

// Realize struct contains the general app informations
type realize struct {
	Name, Description, Author, Email, Host string
	Version                                string
	Limit                                  uint64
	Blueprint                              c.Blueprint
	Server                                 s.Server
	Files                                  map[string]string
	Sync                                   chan string
}

// Application initialization
func init() {
	r = realize{
		Name:        Name,
		Version:     Version,
		Description: Description,
		Host:        Host,
		Limit:       Limit,
		Files: map[string]string{
			"config": Config,
			"output": Output,
		},
		Sync: make(chan string),
	}
	r.Blueprint = c.Blueprint{
		Files: r.Files,
		Sync:  r.Sync,
	}
	r.Server = s.Server{
		Blueprint: &r.Blueprint,
		Files:     r.Files,
		Sync:      r.Sync,
	}
	r.Increase()
	R = &r
}

// Flimit defines the max number of watched files
func (r *realize) Increase() {
	// increases the files limit
	var rLimit syscall.Rlimit
	rLimit.Max = r.Limit
	rLimit.Cur = r.Limit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatal(c.Red("Error Setting Rlimit "), err)
	}
}

func (r *realize) Red(s string) string {
	return c.Red(s)
}

func (r *realize) Blue(s string) string {
	return c.Blue(s)
}

func (r *realize) BlueS(s string) string {
	return c.BlueS(s)
}

func (r *realize) Wdir() string {
	return c.WorkingDir()
}

func (r *realize) Serve(p *cli.Context) {
	if !p.Bool("no-server") {
		fmt.Println(r.Red(r.Host) + "\n")
		r.Server.Open = p.Bool("open")
		r.Server.Start()
	}
}

func (r *realize) Run(p *cli.Context) error {
	r.Serve(p)
	return r.Blueprint.Run()
}

func (r *realize) Fast(p *cli.Context) error {
	r.Blueprint.Add(p)
	r.Serve(p)
	return r.Blueprint.Fast(p)
}

func (r *realize) Add(p *cli.Context) error {
	return r.Blueprint.Insert(p)
}

func (r *realize) Remove(p *cli.Context) error {
	return r.Blueprint.Insert(p)
}

func (r *realize) List(p *cli.Context) error {
	return r.Blueprint.List()
}

func (r *realize) Before(p *cli.Context) error {
	fmt.Println(r.Blue(r.Name) + " - " + r.Blue(r.Version))
	fmt.Println(r.BlueS(r.Description) + "\n")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatal(r.Red("$GOPATH isn't set up properly"))
	}
	return nil
}

func (r *realize) Handle(err error) error {
	if err != nil {
		fmt.Println(r.Red(err.Error()))
		return nil
	}
	return nil
}
