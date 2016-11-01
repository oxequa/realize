package cli

import (
	c "github.com/tockins/realize/settings"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Log struct
type logWriter struct {
	c.Colors
}

// Projects struct contains a projects list
type Blueprint struct {
	*c.Settings `yaml:"-"`
	Projects    []Project   `yaml:"projects,omitempty"`
	Sync        chan string `yaml:"-"`
}

// Project defines the informations of a single project
type Project struct {
	c.Settings    `yaml:"-"`
	LastChangedOn time.Time `yaml:"-"`
	base          string
	Name          string   `yaml:"name,omitempty"`
	Path          string   `yaml:"path,omitempty"`
	Run           bool     `yaml:"run,omitempty"`
	Bin           bool     `yaml:"bin,omitempty"`
	Build         bool     `yaml:"build,omitempty"`
	Fmt           bool     `yaml:"fmt,omitempty"`
	Test          bool     `yaml:"test,omitempty"`
	Params        []string `yaml:"params,omitempty"`
	Watcher       Watcher  `yaml:"watcher,omitempty"`
	Buffer        Buffer   `yaml:"-"`
	parent        *Blueprint
	path          string
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
	StdOut []BufferOut
	StdLog []BufferOut
	StdErr []BufferOut
}

type BufferOut struct {
	Time time.Time
	Text string
}

// Initialize the application
func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
