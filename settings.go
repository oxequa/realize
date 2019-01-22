package core

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

const (
	logs   = "realize.log"
	config = "realize.yaml"
)

// Polling force polling and set a custom interval
type Polling struct {
	Active   bool          `yaml:"active" json:"active"`
	Interval time.Duration `yaml:"interval" json:"interval"`
}

// Settings is a group of general settings
type Settings struct {
	Logs      Logs    `yaml:"logs" json:"logs"`
	Server    Server  `yaml:"server" json:"server"`
	Polling   Polling `yaml:"polling" json:"polling"`
	FileLimit int32   `yaml:"flimit" json:"flimit"`
}

// Broker send information about error
type Logs struct {
	Recovery bool `yaml:"recovery" json:"recovery"`
	File     bool `yaml:"file" json:"file"`
}

// Write config
func (s *Settings) Write(out interface{}) error {
	y, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config, y, 0775)
	if err != nil {
		return err
	}
	return nil
}

// Read config file
func (s *Settings) Read(out interface{}) error {
	// backward compatibility
	if _, err := os.Stat(config); err != nil {
		return err
	}
	content, err := s.Stream(config)
	if err == nil {
		err = yaml.Unmarshal(content, out)
		return err
	}
	return err
}

// Stream return a byte stream of a given file
func (s Settings) Stream(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return content, err
}
