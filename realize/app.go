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
var Red = color.New(color.FgRed, color.Bold).SprintFunc()
var RedS = color.New(color.FgRed).SprintFunc()
var BlueS = color.New(color.FgBlue).SprintFunc()
var Blue = color.New(color.FgBlue, color.Bold).SprintFunc()
var Yellow = color.New(color.FgYellow, color.Bold).SprintFunc()
var MagentaS = color.New(color.FgMagenta).SprintFunc()
var Magenta = color.New(color.FgMagenta, color.Bold).SprintFunc()

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
	fmt.Println(BlueS(app.Description) + "\n")
}
