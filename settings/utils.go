package settings

import (
	"log"
	"os"
	"path/filepath"
)

func (s Settings) Wdir() string {
	dir, err := os.Getwd()
	s.Validate(err)
	return filepath.Base(dir)
}

func (s Settings) Validate(err error) error {
	if err != nil {
		s.Fatal("", err)
	}
	return nil
}

func (s Settings) Fatal(msg string, err error) {
	if msg != "" {
		log.Fatal(s.Red.Regular(msg), err.Error())
	}
	log.Fatal(err.Error())
}
