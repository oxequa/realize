package app

import (
	"fmt"
	w "github.com/tockins/realize/cli"
	c "github.com/tockins/realize/config"
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
	Config      = "R.config.yaml"
	Output      = "R.output.log"
	Host        = "Web server listening on localhost:5000"
)

var R Realize

// Realize struct contains the general app informations
type Realize struct {
	c.Config
	Name, Description, Author, Email, Host string
	Version                                string
	Limit                                  uint64
	Blueprint                              w.Blueprint
	Server                                 s.Server
	Files                                  map[string]string
	Sync                                   chan string
}

// Application initialization
func init() {
	R = Realize{
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
	R.Blueprint = w.Blueprint{
		Config: R.Config,
		Files:  R.Files,
		Sync:   R.Sync,
	}
	R.Server = s.Server{
		Blueprint: &R.Blueprint,
		Files:     R.Files,
		Sync:      R.Sync,
	}
	R.limit()
}

// Flimit defines the max number of watched files
func (r *Realize) limit() {
	var rLimit syscall.Rlimit
	rLimit.Max = R.Limit
	rLimit.Cur = R.Limit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatal(w.Red("Error Setting Rlimit "), err)
	}
}

func (r *Realize) Red(s string) string {
	return w.Red(s)
}

func (r *Realize) Blue(s string) string {
	return w.Blue(s)
}

func (r *Realize) BlueS(s string) string {
	return w.BlueS(s)
}

func (r *Realize) Dir() string {
	return r.Wdir()
}

func (r *Realize) Serve(p *cli.Context) {
	if !p.Bool("no-server") {
		fmt.Println(R.Red(R.Host) + "\n")
		R.Server.Open = p.Bool("open")
		R.Server.Start()
	}
}

func (r *Realize) Run(p *cli.Context) error {
	R.Serve(p)
	return R.Blueprint.Run()
}

func (r *Realize) Fast(p *cli.Context) error {
	R.Blueprint.Add(p)
	R.Serve(p)
	return R.Blueprint.Fast(p)
}

func (r *Realize) Add(p *cli.Context) error {
	return R.Blueprint.Insert(p)
}

func (r *Realize) Remove(p *cli.Context) error {
	return R.Blueprint.Insert(p)
}

func (r *Realize) List(p *cli.Context) error {
	return R.Blueprint.List()
}

func (r *Realize) Before(p *cli.Context) error {
	fmt.Println(R.Blue(R.Name) + " - " + R.Blue(R.Version))
	fmt.Println(R.BlueS(R.Description) + "\n")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatal(R.Red("$GOPATH isn't set up properly"))
	}
	return nil
}

func (r *Realize) Handle(err error) error {
	if err != nil {
		fmt.Println(R.Red(err.Error()))
		return nil
	}
	return nil
}
