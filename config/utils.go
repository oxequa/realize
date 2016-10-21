package config

import (
	"log"
	"os"
	"path/filepath"
)

type Utils struct{}

func Wdir() string {
	dir, err := os.Getwd()
	Validate(err)
	return filepath.Base(dir)
}

func Validate(err error) error {
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
