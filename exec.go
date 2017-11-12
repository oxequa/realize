package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GoCompile is used for compile a project
func (p *Project) goCompile(stop <-chan bool, method []string, args []string) (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan error)
	args = append(method, args...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = p.Path
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
		return msgStop, nil
	case err := <-done:
		// Command completed
		if err != nil {
			return stderr.String(), err
		}
	}
	return "", nil
}

// GoRun  is an implementation of the bin execution
func (p *Project) goRun(stop <-chan bool, runner chan bool) {
	var build *exec.Cmd
	var args []string

	// custom error pattern
	isErrorText := func(string) bool {
		return false
	}
	errRegexp, err := regexp.Compile(p.ErrorOutputPattern)
	if err != nil {
		msg := fmt.Sprintln(p.pname(p.Name, 3), ":", blue.regular(err.Error()))
		out := BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Run"}
		p.stamp("error", out, msg, "")
	} else {
		isErrorText = func(t string) bool {
			return errRegexp.MatchString(t)
		}
	}

	// add additional arguments
	for _, arg := range p.Args {
		a := strings.FieldsFunc(arg, func(i rune) bool {
			return i == '"' || i == '=' || i == '\''
		})
		args = append(args, a...)
	}

	gobin := os.Getenv("GOBIN")
	path := filepath.Join(gobin, p.name)
	if _, err := os.Stat(path); err == nil {
		build = exec.Command(path, args...)
	} else if _, err := os.Stat(path + extWindows); err == nil {
		build = exec.Command(path+extWindows, args...)
	} else {
		path := filepath.Join(p.Path, p.name)
		if _, err = os.Stat(path); err == nil {
			build = exec.Command(path, args...)
		} else if _, err = os.Stat(path + extWindows); err == nil {
			build = exec.Command(path+extWindows, args...)
		} else {
			p.err(errors.New("Build not found"))
			return
		}
	}

	defer func() {
		if err := build.Process.Kill(); err != nil {
			p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Failed to stop: " + err.Error()})
			p.fatal(err, "Failed to stop", ":")
		}
		msg := fmt.Sprintln(p.pname(p.Name, 2), ":", red.regular("Ended"))
		out := BufferOut{Time: time.Now(), Text: "Ended", Type: "Go Run"}
		p.stamp("log", out, msg, "")
	}()

	// scan project stream
	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()
	if err != nil {
		log.Println(red.bold(err.Error()))
		return
	}
	if err := build.Start(); err != nil {
		log.Println(red.bold(err.Error()))
		return
	}
	close(runner)

	execOutput, execError := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOutput, stopError := make(chan bool, 1), make(chan bool, 1)
	scanner := func(stop chan bool, output *bufio.Scanner, isError bool) {
		for output.Scan() {
			text := output.Text()
			msg := fmt.Sprintln(p.pname(p.Name, 3), ":", blue.regular(text))
			if isError && !isErrorText(text) {
				out := BufferOut{Time: time.Now(), Text: text, Type: "Go Run"}
				p.stamp("error", out, msg, "")
			} else {
				out := BufferOut{Time: time.Now(), Text: text, Type: "Go Run"}
				p.stamp("out", out, msg, "")
			}
		}
		close(stop)
	}
	go scanner(stopOutput, execOutput, false)
	go scanner(stopError, execError, true)
	for {
		select {
		case <-stop:
			return
		case <-stopOutput:
			return
		case <-stopError:
			return
		}
	}
}

// Exec an additional command from a defined path if specified
func (p *Project) command(stop <-chan bool, cmd Command) (string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan error)
	args := strings.Split(strings.Replace(strings.Replace(cmd.Command, "'", "", -1), "\"", "", -1), " ")
	ex := exec.Command(args[0], args[1:]...)
	ex.Dir = p.Path
	// make cmd path
	if cmd.Path != "" {
		if strings.Contains(cmd.Path, p.Path) {
			ex.Dir = cmd.Path
		} else {
			ex.Dir = filepath.Join(p.Path, cmd.Path)
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
		return "", ""
	case err := <-done:
		// Command completed
		if err != nil {
			return stderr.String(), stdout.String()
		}
	}
	return "", stdout.String()
}

// GoTool is used for run go tools methods such as fmt, test, generate and so on
func (p *Project) goTool(wg *sync.WaitGroup, stop <-chan bool, result chan<- tool, path string, tool tool) {
	defer wg.Done()
	if tool.status {
		if tool.dir && filepath.Ext(path) != "" {
			path = filepath.Dir(path)
		}
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "") {
			if strings.HasSuffix(path, ".go") {
				tool.options = append(tool.options, path)
				path = p.Path
			}
			if s := ext(path); s == "" || s == "go" {
				var out, stderr bytes.Buffer
				done := make(chan error)
				tool.cmd = append(tool.cmd, tool.options...)
				cmd := exec.Command(tool.cmd[0], tool.cmd[1:]...)
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
					return
				case err := <-done:
					// Command completed
					if err != nil {
						tool.err = stderr.String() + out.String()
						// send command result
						result <- tool
					}
					return
				}

			}
		}
	}
}
