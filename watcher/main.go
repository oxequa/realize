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
	Name          string   `yaml:"name"`
	Path          string   `yaml:"path"`
	Run           bool     `yaml:"run"`
	Bin           bool     `yaml:"bin"`
	Generate      bool     `yaml:"generate"`
	Build         bool     `yaml:"build"`
	Fmt           bool     `yaml:"fmt"`
	Test          bool     `yaml:"test"`
	Params        []string `yaml:"params"`
	Watcher       Watcher  `yaml:"watcher"`
	Cli           Cli      `yaml:"cli"`
	File          File     `yaml:"file"`
	Buffer        Buffer   `yaml:"-"`
	parent        *Blueprint
	path          string
}

// Watcher struct defines the livereload's logic
type Watcher struct {
	// different before and after on re-run?
	Before  []string `yaml:"before"`
	After   []string `yaml:"after"`
	Paths   []string `yaml:"paths"`
	Ignore  []string `yaml:"ignore_paths"`
	Exts    []string `yaml:"exts"`
	Preview bool     `yaml:"preview"`
}

type Cli struct {
	Streams bool `yaml:"streams"`
}

type File struct {
	Streams bool `yaml:"streams"`
	Logs    bool `yaml:"logs"`
	Errors  bool `yaml:"errors"`
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
