package cli

import (
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Read a file given a name and return its byte stream
func read(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err == nil {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		return content, err
	}
	return nil, err

}

// Write a file given a name and a byte stream
func write(name string, data []byte) error {
	err := ioutil.WriteFile(name, data, 0655)
	if err != nil {
		log.Fatal(Red(err))
		return err
	}
	return nil
}

// Create a new file and return its pointer
func create(file string) *os.File {
	out, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0655)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// argsParam parse one by one the given argumentes
func argsParam(params *cli.Context) []string {
	argsN := params.NArg()
	if argsN > 0 {
		var args []string
		for i := 0; i <= argsN-1; i++ {
			args = append(args, params.Args().Get(i))
		}
		return args
	}
	return nil
}

// NameParam check the project name presence. If empty takes the working directory name
func nameFlag(params *cli.Context) string {
	var name string
	if params.String("name") == "" && params.String("path") == "" {
		return App.Wdir()
	} else if params.String("path") != "/" {
		name = filepath.Base(params.String("path"))
	} else {
		name = params.String("name")
	}
	return name
}

// BoolParam is used to check the presence of a bool flag
func boolFlag(b bool) bool {
	if b {
		return false
	}
	return true
}

// Duplicates check projects with same name or same combinations of main/path
func duplicates(value Project, arr []Project) (Project, error) {
	for _, val := range arr {
		if value.Path == val.Path || value.Name == val.Name {
			return val, errors.New("There is a duplicate of '" + val.Name + "'. Check your config file!")
		}
	}
	return Project{}, nil
}

// check if a string is inArray
func inArray(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// Defines the colors scheme for the project name
func pname(name string, color int) string {
	switch color {
	case 1:
		name = Yellow("[") + strings.ToUpper(name) + Yellow("]")
		break
	case 2:
		name = Yellow("[") + Red(strings.ToUpper(name)) + Yellow("]")
		break
	case 3:
		name = Yellow("[") + Blue(strings.ToUpper(name)) + Yellow("]")
		break
	case 4:
		name = Yellow("[") + Magenta(strings.ToUpper(name)) + Yellow("]")
		break
	case 5:
		name = Yellow("[") + Green(strings.ToUpper(name)) + Yellow("]")
		break
	}
	return name
}

// Log struct
type logWriter struct{}

// Cewrites the log timestamp
func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(YellowS("[") + time.Now().Format("15:04:05") + YellowS("]") + string(bytes))
}
