package realize

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"sync"
	"time"
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
var YellowS = color.New(color.FgYellow).SprintFunc()
var MagentaS = color.New(color.FgMagenta).SprintFunc()
var Magenta = color.New(color.FgMagenta, color.Bold).SprintFunc()

var watcherIgnores = []string{"vendor", "bin"}
var watcherExts = []string{".go"}
var watcherPaths = []string{"/"}

type logWriter struct{}

// App struct contains the informations about realize
type App struct {
	Name, Version, Description, Author, Email string
}

// Custom log timestamp
func init() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
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

// Cewrites the log timestamp
func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(YellowS("[") + time.Now().UTC().Format("15:04:05") + YellowS("]") + string(bytes))
}
