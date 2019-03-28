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
	Sync     chan string    `yaml:"-" json:"-"`
	Exit     chan os.Signal `yaml:"-" json:"-"`
	Settings Settings       `yaml:"settings" json:"settings"`
	Projects []Project      `yaml:"projects,omitempty" json:"projects,omitempty"`
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

// TempFile check if a given filepath is a temp file
func TempFile(path string) bool {
	ext := filepath.Ext(path)
	baseName := filepath.Base(path)
	temp := strings.HasSuffix(ext, "~") ||
		(ext == ".swp") || // vim
		(ext == ".swx") || // vim
		(ext == ".tmp") || // generic temp file
		(ext == ".DS_Store") || // OSX Thumbnail
		baseName == "4913" || // vim
		strings.HasPrefix(ext, ".goutputstream") || // gnome
		strings.HasSuffix(ext, "jb_old___") || // intelliJ
		strings.HasSuffix(ext, "jb_tmp___") || // intelliJ
		strings.HasSuffix(ext, "jb_bak___") || // intelliJ
		strings.HasPrefix(ext, ".sb-") || // byword
		strings.HasPrefix(baseName, ".#") || // emacs
		strings.HasPrefix(baseName, "#") // emacs
	return temp
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

// Check in slice a given string
func CheckInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
