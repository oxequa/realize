package core

import (
	"bytes"
	"encoding/gob"
	"github.com/getsentry/raven-go"
	"io/ioutil"
	"log"
	"time"
)

const (
	perm     = 0775
	recovery = "recovery.log"
)

// Legacy force polling and set a custom interval
type Legacy struct {
	Force    bool          `yaml:"force,omitempty" json:"force,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
}

// Options is a group of general settings
type Options struct {
	FileLimit int32             `yaml:"flimit,omitempty" json:"flimit,omitempty"`
	Legacy    Legacy            `yaml:"legacy,omitempty" json:"legacy,omitempty"`
	Broker    Broker            `yaml:"broker,omitempty" json:"broker,omitempty"`
	Env       map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

// Broker send informations about error
type Broker struct {
	Log  bool `yaml:"log,omitempty" json:"log,omitempty"`
	File bool `yaml:"file,omitempty" json:"file,omitempty"`
	Send bool `yaml:"send,omitempty" json:"send,omitempty"`
}

// Push a new msg on recovery cloud, file or cli if are enabled. It is useful to debug realize flow.
func (r *Broker) Push(format string, m interface{}) (e error) {

	// switch type

	if m != nil {
		if r.Send {
			switch t := m.(type) {
			case error:
				raven.CaptureError(t, nil)
			}
		}
		if r.Log {
			log.Println(format, m)
		}
		if r.File {
			var buf bytes.Buffer
			e = gob.NewEncoder(&buf).Encode(m)
			if e != nil {
				r.File = false
				return r.Push(Prefix("Error", Red), m)
			}
			e = ioutil.WriteFile(recovery, buf.Bytes(), perm)
			if e != nil {
				r.File = false
				return r.Push(Prefix("Error", Red), m)
			}
		}
	}
	return
}
