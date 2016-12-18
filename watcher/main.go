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

// Blueprint struct contains a projects list
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
	Fmt           bool     `yaml:"fmt" json:"fmt"`
	Test          bool     `yaml:"test" json:"test"`
	Generate      bool     `yaml:"generate" json:"generate"`
	Bin           bool     `yaml:"bin" json:"bin"`
	Build         bool     `yaml:"build" json:"build"`
	Run           bool     `yaml:"run" json:"run"`
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

// Cli output status, enables or disables
type Cli struct {
	Streams bool `yaml:"streams" json:"streams"`
}

// File determinates the status of each log files (streams, logs, errors)
type File struct {
	Streams bool `yaml:"streams" json:"streams"`
	Logs    bool `yaml:"logs" json:"logs"`
	Errors  bool `yaml:"errors" json:"errors"`
}

// Buffer define an array buffer for each log files
type Buffer struct {
	StdOut []BufferOut `json:"stdOut"`
	StdLog []BufferOut `json:"stdLog"`
	StdErr []BufferOut `json:"stdErr"`
}

// BufferOut is used for exchange information between "realize cli" and "web realize"
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
