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
	Projects    []Project   `yaml:"projects,omitempty" json:"projects,omitempty"`
	Sync        chan string `yaml:"-"`
}

// Project defines the informations of a single project
type Project struct {
	c.Settings    `yaml:"-"`
	LastChangedOn time.Time `yaml:"-" json:"-"`
	base          string
	Name          string   `yaml:"name" json:"name"`
	Path          string   `yaml:"path" json:"path"`
	Run           bool     `yaml:"run" json:"run"`
	Bin           bool     `yaml:"bin" json:"bin"`
	Generate      bool     `yaml:"generate" json:"generate"`
	Build         bool     `yaml:"build" json:"build"`
	Fmt           bool     `yaml:"fmt" json:"fmt"`
	Test          bool     `yaml:"test" json:"test"`
	Params        []string `yaml:"params" json:"params"`
	Watcher       Watcher  `yaml:"watcher" json:"watcher"`
	Cli           Cli      `yaml:"cli" json:"cli"`
	File          File     `yaml:"file" json:"file"`
	Buffer        Buffer   `yaml:"-" json:"buffer"`
	parent        *Blueprint
	path          string
}

// Watcher struct defines the livereload's logic
type Watcher struct {
	// different before and after on re-run?
	Before  []string `yaml:"before" json:"before"`
	After   []string `yaml:"after" json:"after"`
	Paths   []string `yaml:"paths" json:"paths"`
	Ignore  []string `yaml:"ignore_paths" json:"ignore"`
	Exts    []string `yaml:"exts" json:"exts"`
	Preview bool     `yaml:"preview" json:"preview"`
}

type Cli struct {
	Streams bool `yaml:"streams" json:"streams"`
}

type File struct {
	Streams bool `yaml:"streams" json:"streams"`
	Logs    bool `yaml:"logs" json:"logs"`
	Errors  bool `yaml:"errors" json:"errors"`
}

// Buffer struct for buffering outputs
type Buffer struct {
	StdOut []BufferOut `json:"stdOut"`
	StdLog []BufferOut `json:"stdLog"`
	StdErr []BufferOut `json:"stdErr"`
}

type BufferOut struct {
	Time   time.Time `json:"time"`
	Text   string    `json:"text"`
	Path   string    `json:"path"`
	Type   string    `json:"type"`
	Stream string    `json:"stream"`
	Errors []string  `json:"errors"`
}

// Initialize the application
func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
