package cli

import (
	"github.com/fatih/color"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup

var Green, Red, RedS, Blue, BlueS, Yellow, YellowS, Magenta, MagentaS = color.New(color.FgGreen, color.Bold).SprintFunc(),
	color.New(color.FgRed, color.Bold).SprintFunc(),
	color.New(color.FgRed).SprintFunc(),
	color.New(color.FgBlue, color.Bold).SprintFunc(),
	color.New(color.FgBlue).SprintFunc(),
	color.New(color.FgYellow, color.Bold).SprintFunc(),
	color.New(color.FgYellow).SprintFunc(),
	color.New(color.FgMagenta, color.Bold).SprintFunc(),
	color.New(color.FgMagenta).SprintFunc()

// Log struct
type logWriter struct{}

// Projects struct contains a projects list
type Blueprint struct {
	Projects []Project         `yaml:"projects,omitempty"`
	Files    map[string]string `yaml:"-"`
	Sync     chan string
}

// Project defines the informations of a single project
type Project struct {
	reload  time.Time
	base    string
	Name    string   `yaml:"app_name,omitempty"`
	Path    string   `yaml:"app_path,omitempty"`
	Run     bool     `yaml:"app_run,omitempty"`
	Bin     bool     `yaml:"app_bin,omitempty"`
	Build   bool     `yaml:"app_build,omitempty"`
	Fmt     bool     `yaml:"app_fmt,omitempty"`
	Test    bool     `yaml:"app_test,omitempty"`
	Params  []string `yaml:"app_params,omitempty"`
	Watcher Watcher  `yaml:"app_watcher,omitempty"`
	Buffer  Buffer   `yaml:"-"`
	parent  *Blueprint
}

// Watcher struct defines the livereload's logic
type Watcher struct {
	// different before and after on re-run?
	Before  []string        `yaml:"before,omitempty"`
	After   []string        `yaml:"after,omitempty"`
	Paths   []string        `yaml:"paths,omitempty"`
	Ignore  []string        `yaml:"ignore_paths,omitempty"`
	Exts    []string        `yaml:"exts,omitempty"`
	Preview bool            `yaml:"preview,omitempty"`
	Output  map[string]bool `yaml:"output,omitempty"`
}

// Buffer struct for buffering outputs
type Buffer struct {
	StdOut []string
	StdLog []string
	StdErr []string
}

// Initialize the application
func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
