package realize

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool info
type Tool struct {
	Args   []string `yaml:"args,omitempty" json:"args,omitempty"`
	Method string   `yaml:"method,omitempty" json:"method,omitempty"`
	Status bool     `yaml:"status,omitempty" json:"status,omitempty"`
	dir    bool
	isTool bool
	method []string
	cmd    []string
	name   string
}

// Tools go
type Tools struct {
	Fix      Tool `yaml:"fix,omitempty" json:"fix,omitempty"`
	Clean    Tool `yaml:"clean,omitempty" json:"clean,omitempty"`
	Vet      Tool `yaml:"vet,omitempty" json:"vet,omitempty"`
	Fmt      Tool `yaml:"fmt,omitempty" json:"fmt,omitempty"`
	Test     Tool `yaml:"test,omitempty" json:"test,omitempty"`
	Generate Tool `yaml:"generate,omitempty" json:"generate,omitempty"`
	Install  Tool `yaml:"install,omitempty" json:"install,omitempty"`
	Build    Tool `yaml:"build,omitempty" json:"build,omitempty"`
	Run      bool `yaml:"run,omitempty" json:"run,omitempty"`
}

// Setup go tools
func (t *Tools) Setup() {
	// go clean
	if t.Clean.Status {
		t.Clean.name = "Clean"
		t.Clean.isTool = true
		t.Clean.cmd = replace([]string{"go clean"}, t.Clean.Method)
		t.Clean.Args = split([]string{}, t.Clean.Args)
	}
	// go generate
	if t.Generate.Status {
		t.Generate.dir = true
		t.Generate.isTool = true
		t.Generate.name = "Generate"
		t.Generate.cmd = replace([]string{"go", "generate"}, t.Generate.Method)
		t.Generate.Args = split([]string{}, t.Generate.Args)
	}
	// go fix
	if t.Fix.Status {
		t.Fix.name = "Fix"
		t.Fix.isTool = true
		t.Fix.cmd = replace([]string{"go fix"}, t.Fix.Method)
		t.Fix.Args = split([]string{}, t.Fix.Args)
	}
	// go fmt
	if t.Fmt.Status {
		if len(t.Fmt.Args) == 0 {
			t.Fmt.Args = []string{"-s", "-w", "-e", "./"}
		}
		t.Fmt.name = "Fmt"
		t.Fmt.isTool = true
		t.Fmt.cmd = replace([]string{"gofmt"}, t.Fmt.Method)
		t.Fmt.Args = split([]string{}, t.Fmt.Args)
	}
	// go vet
	if t.Vet.Status {
		t.Vet.dir = true
		t.Vet.name = "Vet"
		t.Vet.isTool = true
		t.Vet.cmd = replace([]string{"go", "vet"}, t.Vet.Method)
		t.Vet.Args = split([]string{}, t.Vet.Args)
	}
	// go test
	if t.Test.Status {
		t.Test.dir = true
		t.Test.isTool = true
		t.Test.name = "Test"
		t.Test.cmd = replace([]string{"go", "test"}, t.Test.Method)
		t.Test.Args = split([]string{}, t.Test.Args)
	}
	// go install
	t.Install.name = "Install"
	t.Install.cmd = replace([]string{"go", "install"}, t.Install.Method)
	t.Install.Args = split([]string{}, t.Install.Args)
	// go build
	if t.Build.Status {
		t.Build.name = "Build"
		t.Build.cmd = replace([]string{"go", "build"}, t.Build.Method)
		t.Build.Args = split([]string{}, t.Build.Args)
	}
}

// Exec a go tool
func (t *Tool) Exec(path string, stop <-chan bool) (response Response) {
	if t.dir && filepath.Ext(path) != "" {
		path = filepath.Dir(path)
	}
	if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "") {
		args := []string{}
		if strings.HasSuffix(path, ".go") {
			args = append(t.Args, path)
			path = filepath.Dir(path)
		}
		if s := ext(path); s == "" || s == "go" {
			var out, stderr bytes.Buffer
			done := make(chan error)
			args = append(t.cmd, t.Args...)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = path
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			// Start command
			cmd.Start()
			go func() { done <- cmd.Wait() }()
			// Wait a result
			select {
			case <-stop:
				// Stop running command
				cmd.Process.Kill()
			case err := <-done:
				// Command completed
				response.Name = t.name
				if err != nil {
					response.Err = errors.New(stderr.String() + out.String())
				} else {
					response.Out = out.String()
				}
			}
		}
	}
	return
}

// Compile is used for build and install
func (t *Tool) Compile(path string, stop <-chan bool) (response Response) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan error)
	args := append(t.cmd, t.Args...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(path)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	// Start command
	cmd.Start()
	go func() { done <- cmd.Wait() }()
	// Wait a result
	select {
	case <-stop:
		// Stop running command
		cmd.Process.Kill()
	case err := <-done:
		// Command completed
		response.Name = t.name
		if err != nil {
			response.Err = errors.New(stderr.String() + err.Error())
		}
	}
	return
}
