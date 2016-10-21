package config

import (
	"log"
	"os"
	"path/filepath"
)

type Utils struct{}

func (u *Utils) Wdir() string {
	dir, err := os.Getwd()
	u.Validate(err)
	return filepath.Base(dir)
}

func (u *Utils) Validate(err error) error {
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
