package realize

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"sync"
	"syscall"
	"time"
)

// Default values and info
const (
	AppName        = "Realize"
	AppVersion     = "v1.0"
	AppEmail       = "pracchia@hastega.it"
	AppDescription = "Run, install or build your applications on file changes. Output preview and multi project support"
	AppAuthor      = "Alessio Pracchia"
	AppFile        = "realize.config.yaml"
)

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

	// increases the files limit
	var rLimit syscall.Rlimit
	rLimit.Max = 10000
	rLimit.Cur = 10000
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println(Red("Error Setting Rlimit "), err)
	}
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
