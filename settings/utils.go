package settings

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tockins/realize/style"
)

// Wdir return the current working Directory
func (s Settings) Wdir() string {
	dir, err := os.Getwd()
	s.Validate(err)
	return filepath.Base(dir)
}

// Validate checks a fatal error
func (s Settings) Validate(err error) error {
	if err != nil {
		s.Fatal(err, "")
	}
	return nil
}

// Fatal prints a fatal error with its additional messages
func (s Settings) Fatal(err error, msg ...interface{}) {
	if len(msg) > 0 && err != nil {
		log.Fatalln(style.Red.Regular(msg...), err.Error())
	} else if err != nil {
		log.Fatalln(err.Error())
	}
}

// Name return the project name or the path of the working dir
func (s Settings) Name(name string, path string) string {
	if name == "" && path == "" {
		return s.Wdir()
	} else if path != "/" {
		return filepath.Base(path)
	}
	return name
}

// Path cleaner
func (s Settings) Path(path string) string {
	return strings.Replace(filepath.Clean(path), "\\", "/", -1)
}
