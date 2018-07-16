package core

import (
	"bytes"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Custom log
type Log struct{}

// Realize main struct
type Realize struct {
	Sync    chan string `yaml:"-" json:"-"`
	Exit    chan bool   `yaml:"-" json:"-"`
	Server  Server      `yaml:"server,omitempty" json:"server,omitempty"`
	Options Options     `yaml:"options,omitempty" json:"options,omitempty"`
	Schema  []Activity  `yaml:"schema,inline,omitempty" json:"schema,inline,omitempty"`
}

// initial set up
func init() {
	// custom log
	log.SetFlags(0)
	log.SetOutput(Log{})
	if build.Default.GOPATH == "" {
		log.Fatal("GOPATH isn't set properly")
	}
	path := filepath.SplitList(build.Default.GOPATH)
	if err := os.Setenv("GOBIN", filepath.Join(path[len(path)-1], "bin")); err != nil {
		log.Fatal("GOBIN impossible to set", err)
	}
}

// Ext return file extension
func Ext(path string) string {
	var ext string
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			ext = path[i:]
			if index := strings.LastIndex(ext, "."); index > 0 {
				ext = ext[index:]
			}
		}
	}
	if ext != "" {
		return ext[1:]
	}
	return ""
}

// Hidden check an hidden file
func Hidden(path string) bool {
	if runtime.GOOS != "windows" {
		if filepath.Base(path)[0:1] == "." {
			return true
		}
	}
	return false
	// need a way to check on windows
}

func Print(msg ...interface{}) string {
	var buffer bytes.Buffer
	for i := 0; i < len(msg); i++ {
		buffer.WriteString(fmt.Sprint(msg[i]) + " ")
	}
	return buffer.String()
}

// Rewrite timestamp log layout
func (l Log) Write(bytes []byte) (int, error) {
	if len(bytes) > 0 {
		return fmt.Fprint(Output, Yellow.Regular("["), time.Now().Format("15:04:05"), Yellow.Regular("]"), string(bytes))
	}
	return 0, nil
}
