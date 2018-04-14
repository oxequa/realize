package realize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/go-siris/siris/core/errors"
	"go/build"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	// RPrefix tool name
	RPrefix = "realize"
	// RVersion current version
	RVersion = "2.1"
	// RExt file extension
	RExt = ".yaml"
	// RFile config file name
	RFile = "." + RPrefix + RExt
	//RExtWin windows extension
	RExtWin = ".exe"
)

type (
	// LogWriter used for all log
	LogWriter struct{}

	// Realize main struct
	Realize struct {
		Settings Settings `yaml:"settings" json:"settings"`
		Server   Server   `yaml:"server,omitempty" json:"server,omitempty"`
		Schema   `yaml:",inline" json:",inline"`
		Sync     chan string `yaml:"-" json:"-"`
		Err      Func        `yaml:"-" json:"-"`
		After    Func        `yaml:"-"  json:"-"`
		Before   Func        `yaml:"-"  json:"-"`
		Change   Func        `yaml:"-"  json:"-"`
		Reload   Func        `yaml:"-"  json:"-"`
	}

	// Context is used as argument for func
	Context struct {
		Path    string
		Project *Project
		Stop    <-chan bool
		Watcher FileWatcher
		Event   fsnotify.Event
	}

	// Func is used instead realize func
	Func func(Context)
)

// init check
func init() {
	// custom log
	log.SetFlags(0)
	log.SetOutput(LogWriter{})
	if build.Default.GOPATH == "" {
		log.Fatal("$GOPATH isn't set properly")
	}
	path := filepath.SplitList(build.Default.GOPATH)
	if err := os.Setenv("GOBIN", filepath.Join(path[len(path)-1], "bin")); err != nil {
		log.Fatal(err)
	}
}

// Stop realize workflow
func (r *Realize) Stop() error {
	for k := range r.Schema.Projects {
		if r.Schema.Projects[k].exit != nil {
			close(r.Schema.Projects[k].exit)
		}
	}
	return nil
}

// Start realize workflow
func (r *Realize) Start() error {
	if len(r.Schema.Projects) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(r.Schema.Projects))
		for k := range r.Schema.Projects {
			r.Schema.Projects[k].exit = make(chan os.Signal, 1)
			signal.Notify(r.Schema.Projects[k].exit, os.Interrupt)
			r.Schema.Projects[k].parent = r
			go r.Schema.Projects[k].Watch(&wg)
		}
		wg.Wait()
	} else {
		return errors.New("there are no projects")
	}
	return nil
}

// Prefix a given string with tool name
func (r *Realize) Prefix(input string) string {
	if len(input) > 0 {
		return fmt.Sprint(Yellow.Bold("["), strings.ToUpper(RPrefix), Yellow.Bold("]"), " : ", input)
	}
	return input
}

// Rewrite the layout of the log timestamp
func (w LogWriter) Write(bytes []byte) (int, error) {
	if len(bytes) > 0 {
		return fmt.Fprint(Output, Yellow.Regular("["), time.Now().Format("15:04:05"), Yellow.Regular("]"), string(bytes))
	}
	return 0, nil
}
