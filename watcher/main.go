package watcher

import (
	"github.com/tockins/realize/settings"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup

// Watcher interface used by polling/fsnotify watching
type watcher interface {
	Add(path string) error
}

// Polling watcher
type pollWatcher struct {
	paths map[string]bool
}

// Log struct
type logWriter struct{}

// Blueprint struct contains a projects list
type Blueprint struct {
	*settings.Settings `yaml:"-"`
	Projects           []Project   `yaml:"projects,omitempty" json:"projects,omitempty"`
	Sync               chan string `yaml:"-"`
}

// Project defines the informations of a single project
type Project struct {
	settings.Settings  `yaml:"-"`
	LastChangedOn      time.Time `yaml:"-" json:"-"`
	base               string
	Name               string            `yaml:"name" json:"name"`
	Path               string            `yaml:"path" json:"path"`
	Environment        map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Cmds               Cmds              `yaml:"commands" json:"commands"`
	Args               []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Watcher            Watcher           `yaml:"watcher" json:"watcher"`
	Streams            Streams           `yaml:"streams,omitempty" json:"streams,omitempty"`
	Buffer             Buffer            `yaml:"-" json:"buffer"`
	ErrorOutputPattern string            `yaml:"errorOutputPattern,omitempty" json:"errorOutputPattern,omitempty"`
	parent             *Blueprint
	path               string
	tools              tools
}

type tools struct {
	Fmt, Test, Generate, Vet tool
}

type tool struct {
	status  *bool
	cmd     string
	options []string
	name    string
}

type Cmds struct {
	Vet      bool `yaml:"vet" json:"vet"`
	Fmt      bool `yaml:"fmt" json:"fmt"`
	Test     bool `yaml:"test" json:"test"`
	Generate bool `yaml:"generate" json:"generate"`
	Bin      Cmd  `yaml:"bin" json:"bin"`
	Build    Cmd  `yaml:"build" json:"build"`
	Run      bool `yaml:"run" json:"run"`
}

// Buildmode options
type Cmd struct {
	Status bool     `yaml:"status" json:"status"`
	Args   []string `yaml:"args,omitempty" json:"args,omitempty"`
}

// Watcher struct defines the livereload's logic
type Watcher struct {
	Preview bool      `yaml:"preview" json:"preview"`
	Paths   []string  `yaml:"paths" json:"paths"`
	Ignore  []string  `yaml:"ignore_paths" json:"ignore"`
	Exts    []string  `yaml:"exts" json:"exts"`
	Scripts []Command `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// Command options
type Command struct {
	Type    string `yaml:"type" json:"type"`
	Command string `yaml:"command" json:"command"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
	Changed bool   `yaml:"changed,omitempty" json:"changed,omitempty"`
	Startup bool   `yaml:"startup,omitempty" json:"startup,omitempty"`
}

// Streams is a collection of names and values for the logs functionality
type Streams struct {
	FileOut bool `yaml:"file_out" json:"file_out"`
	FileLog bool `yaml:"file_log" json:"file_log"`
	FileErr bool `yaml:"file_err" json:"file_err"`
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
