package realize

import (
	"github.com/fatih/color"
	"log"
	"sync"
)

var App Realize

var wg sync.WaitGroup

// Green color bold
var Green = color.New(color.FgGreen, color.Bold).SprintFunc()

// Red color bold
var Red = color.New(color.FgRed, color.Bold).SprintFunc()

// RedS color used for errors
var RedS = color.New(color.FgRed).SprintFunc()

// Blue color bold used for project output
var Blue = color.New(color.FgBlue, color.Bold).SprintFunc()

// BlueS color
var BlueS = color.New(color.FgBlue).SprintFunc()

// Yellow color bold
var Yellow = color.New(color.FgYellow, color.Bold).SprintFunc()

// YellowS color
var YellowS = color.New(color.FgYellow).SprintFunc()

// MagentaS color
var MagentaS = color.New(color.FgMagenta).SprintFunc()

// Magenta color bold
var Magenta = color.New(color.FgMagenta, color.Bold).SprintFunc()

// Initialize the application
func init() {
	App = Realize{
		Name:        "Realize",
		Version:     "1.0",
		Description: "A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths",
		Limit:       10000,
	}
	App.Blueprint.files = map[string]string{
		"config": "r.config.yaml",
		"output": "r.output.log",
	}
	App.limit()
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
