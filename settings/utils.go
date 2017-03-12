package settings

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Wdir return the current working directory
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
		log.Fatalln(s.Red.Regular(msg...), err.Error())
	} else if err != nil {
		log.Fatalln(err.Error())
	}
}

func (h Settings) Name(name string, path string) string {
	if name == "" && path == "" {
		return h.Wdir()
	} else if path != "/" {
		return filepath.Base(path)
	}
	return name
}

func (h Settings) Path(s string) string {
	return strings.Replace(filepath.Clean(s), "\\", "/", -1)
}
