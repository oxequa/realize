package config

import (
	"log"
	"os"
	"path/filepath"
)

func (c *Config) Wdir() string {
	dir, err := os.Getwd()
	c.Validate(err)
	return filepath.Base(dir)
}

func (c *Config) Validate(err error) error {
	if err != nil {
		log.Fatal(Red(err))
	}
	return nil
}

func (c *Config) Fatal(msg string, err error){
	if(msg != "") {
		log.Fatal(Red(msg), err.Error())
	}
	log.Fatal(err.Error())
}
