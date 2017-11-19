package main

import (
	"bytes"
	"github.com/go-siris/siris/core/errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// Command options
type Command struct {
	Type   string `yaml:"type" json:"type"`
	Cmd    string `yaml:"command" json:"command"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Global bool   `yaml:"global,omitempty" json:"global,omitempty"`
	Output bool   `yaml:"output,omitempty" json:"output,omitempty"`
}

// Exec an additional command from a defined path if specified
func (c *Command) Exec(base string, stop <-chan bool) (response Response) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan error)
	args := strings.Split(strings.Replace(strings.Replace(c.Cmd, "'", "", -1), "\"", "", -1), " ")
	ex := exec.Command(args[0], args[1:]...)
	ex.Dir = base
	// make cmd path
	if c.Path != "" {
		if strings.Contains(c.Path, base) {
			ex.Dir = c.Path
		} else {
			ex.Dir = filepath.Join(base, c.Path)
		}
	}
	ex.Stdout = &stdout
	ex.Stderr = &stderr
	// Start command
	ex.Start()
	go func() { done <- ex.Wait() }()
	// Wait a result
	select {
	case <-stop:
		// Stop running command
		ex.Process.Kill()
	case err := <-done:
		// Command completed
		response.Name = c.Cmd
		response.Out = stdout.String()
		if err != nil {
			response.Err = errors.New(stderr.String())
		}
	}
	return
}
