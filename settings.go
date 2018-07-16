package core

import (
	"time"
)

const (
	logFile = "realize.log"
)

// Polling force polling and set a custom interval
type Polling struct {
	Force    bool          `yaml:"force,omitempty" json:"force,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
}

// Settings is a group of general settings
type Settings struct {
	Broker    Broker  `yaml:"broker,omitempty" json:"broker,omitempty"`
	Server    Server  `yaml:"server,omitempty" json:"server,omitempty"`
	Polling   Polling `yaml:"polling,omitempty" json:"polling,omitempty"`
	FileLimit int32   `yaml:"file_limit,omitempty" json:"file_limit,omitempty"`
}

// Broker send information about error
type Broker struct {
	Recovery bool `yaml:"recovery,omitempty" json:"recovery,omitempty"`
	File     bool `yaml:"file,omitempty" json:"file,omitempty"`
}
