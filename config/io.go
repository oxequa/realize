package config

import (
	"io/ioutil"
	"os"
)

// Scan return a byte stream of a given file
func (c *Config) Stream(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err == nil {
		content, err := ioutil.ReadFile(file)
		c.Validate(err)
		return content, err
	}
	return nil, err
}

// Write a file given a name and a byte stream
func (c *Config) Write(name string, data []byte) error {
	err := ioutil.WriteFile(name, data, 0655)
	return c.Validate(err)
}

// Create a new file and return its pointer
func (c *Config) Create(file string) *os.File {
	out, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0655)
	c.Validate(err)
	return out
}
