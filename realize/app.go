package realize

import (
	"fmt"
	"github.com/fatih/color"
	"sync"
)

const (
	AppName        = "Realize"
	AppVersion     = "v1.0"
	AppEmail       = "pracchia@hastega.it"
	AppDescription = "Run, install or build your applications on file changes. Output preview and multi project support"
	AppAuthor      = "Alessio Pracchia"
	AppFile        = "realize.config.yaml"
)

var wg sync.WaitGroup
var Green = color.New(color.FgGreen, color.Bold).SprintFunc()
var Greenl = color.New(color.FgGreen).SprintFunc()
var Red = color.New(color.FgRed, color.Bold).SprintFunc()
var Redl = color.New(color.FgRed).SprintFunc()
var Blue = color.New(color.FgBlue, color.Bold).SprintFunc()
var Bluel = color.New(color.FgBlue).SprintFunc()
var Magenta = color.New(color.FgMagenta).SprintFunc()

var watcherIgnores = []string{"vendor", "bin"}
var watcherExts = []string{".go"}
var watcherPaths = []string{"/"}

// App struct contains the informations about realize
type App struct {
	Name, Version, Description, Author, Email string
}

// Init is an instance of app with default values
func Init() *App {
	return &App{
		Name:        AppName,
		Version:     AppVersion,
		Description: AppDescription,
		Author:      AppAuthor,
		Email:       AppEmail,
	}
}

// Information print realize name and description
func (app *App) Information() {
	fmt.Println(Blue(app.Name) + " - " + Blue(app.Version))
	fmt.Println(Bluel(app.Description) + "\n")
}
