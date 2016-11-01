package settings

import (
	"io/ioutil"
	"os"
)

// Scan return a byte stream of a given file
func (s Settings) Stream(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err == nil {
		content, err := ioutil.ReadFile(file)
		s.Validate(err)
		return content, err
	}
	return nil, err
}

// Write a file given a name and a byte stream
func (s Settings) Write(name string, data []byte) error {
	err := ioutil.WriteFile(name, data, 0655)
	return s.Validate(err)
}

// Create a new file and return its pointer
func (s Settings) Create(file string) *os.File {
	out, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0655)
	s.Validate(err)
	return out
}
