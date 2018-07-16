package core

import (
	"time"
)

const (
	perm    = 0775
	logFile = "realize.log"
)

// Legacy force polling and set a custom interval
type Legacy struct {
	Force    bool          `yaml:"force,omitempty" json:"force,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
}

// Options is a group of general settings
type Options struct {
	FileLimit int32  `yaml:"flimit,omitempty" json:"flimit,omitempty"`
	Legacy    Legacy `yaml:"legacy,omitempty" json:"legacy,omitempty"`
	Broker    Broker `yaml:"broker,omitempty" json:"broker,omitempty"`
}

// Broker send informations about error
type Broker struct {
	Recovery bool `yaml:"recovery,omitempty" json:"recovery,omitempty"`
	File     bool `yaml:"file,omitempty" json:"file,omitempty"`
}
