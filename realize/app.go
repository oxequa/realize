package realize

import (
	"fmt"
	"github.com/fatih/color"
	"log"
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
var green = color.New(color.FgGreen, color.Bold).SprintFunc()
var greenl = color.New(color.FgHiGreen).SprintFunc()
var red = color.New(color.FgRed, color.Bold).SprintFunc()
var blue = color.New(color.FgBlue, color.Bold).SprintFunc()
var bluel = color.New(color.FgBlue).SprintFunc()

var watcherIgnores = []string{"vendor", "bin"}
var watcherExts = []string{".go"}
var watcherPaths = []string{"/"}

// App struct contains the informations about realize
type App struct {
	Name, Version, Description, Author, Email string
}

// Init is an instance of app with default values
func Init() App {
	return App{
		Name:        AppName,
		Version:     AppVersion,
		Description: AppDescription,
		Author:      AppAuthor,
		Email:       AppEmail,
	}
}

// Fail is a red message, generally used for errors
func Fail(msg string) {
	fmt.Println(red(msg))
}

// Success is a green message, generally used for feedback
func Success(msg string) {
	fmt.Println(green(msg))
}

// LogSuccess is a green log message, generally used for feedback
func LogSuccess(msg string) {
	log.Println(green(msg))
}

// LogFail is a red log message, generally used for errors
func LogFail(msg string) {
	log.Println(red(msg))
}

// Information print realize name and description
func (app *App) Information() {
	fmt.Println(blue(app.Name) + " - " + blue(app.Version))
	fmt.Println(bluel(app.Description) + "\n")
}
