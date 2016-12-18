package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GoRun  is an implementation of the bin execution
func (p *Project) goRun(channel chan bool, runner chan bool, wr *sync.WaitGroup) error {
	var build *exec.Cmd
	var params []string
	var path = ""
	for _, param := range p.Params {
		arr := strings.Fields(param)
		params = append(params, arr...)
	}
	if _, err := os.Stat(filepath.Join(p.base, p.path)); err == nil {
		path = filepath.Join(p.base, p.path)
	}
	if _, err := os.Stat(filepath.Join(p.base, p.path+".exe")); err == nil {
		path = filepath.Join(p.base, p.path+".exe")
	}

	if path != "" {
		build = exec.Command(path, params...)
	} else {
		if _, err := os.Stat(filepath.Join(os.Getenv("GOBIN"), filepath.Base(p.path))); err == nil {
			build = exec.Command(filepath.Join(os.Getenv("GOBIN"), filepath.Base(p.path)), params...)
		} else {
			p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Can't run a not compiled project"})
			p.Fatal(err, "Can't run a not compiled project", ":")
		}
	}
	defer func() {
		if err := build.Process.Kill(); err != nil {
			p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Failed to stop: " + err.Error()})
			p.Fatal(err, "Failed to stop", ":")
		}
		msg := fmt.Sprintln(p.pname(p.Name, 2), ":", p.Red.Regular("Ended"))
		out := BufferOut{Time: time.Now(), Text: "Ended", Type: "Go Run"}
		p.print("log", out, msg, "")
		wr.Done()
	}()

	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()
	if err != nil {
		log.Println(p.Red.Bold(err.Error()))
		return err
	}
	if err := build.Start(); err != nil {
		log.Println(p.Red.Bold(err.Error()))
		return err
	}
	close(runner)

	execOutput, execError := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOutput, stopError := make(chan bool, 1), make(chan bool, 1)
	scanner := func(stop chan bool, output *bufio.Scanner, isError bool) {
		for output.Scan() {
			select {
			default:
				msg := fmt.Sprintln(p.pname(p.Name, 3), ":", p.Blue.Regular(output.Text()))
				if isError {
					out := BufferOut{Time: time.Now(), Text: output.Text(), Type: "Go Run"}
					p.print("error", out, msg, "")
				} else {
					out := BufferOut{Time: time.Now(), Text: output.Text()}
					p.print("out", out, msg, "")
				}
			}
		}
		close(stop)
	}
	go scanner(stopOutput, execOutput, false)
	go scanner(stopError, execError, true)
	for {
		select {
		case <-channel:
			return nil
		case <-stopOutput:
			return nil
		case <-stopError:
			return nil
		}
	}
}

// GoBuild is an implementation of the "go build"
func (p *Project) goBuild() (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	build := exec.Command("go", "build")
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// GoInstall is an implementation of the "go install"
func (p *Project) goInstall() (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	err := os.Setenv("GOBIN", filepath.Join(os.Getenv("GOPATH"), "bin"))
	if err != nil {
		return "", err
	}
	build := exec.Command("go", "install")
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// GoTools is used for run go methods such as fmt, test, generate...
func (p *Project) goTools(dir string, name string, cmd ...string) (string, error) {
	var out, stderr bytes.Buffer
	build := exec.Command(name, cmd...)
	build.Dir = dir
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// Cmds exec a list of defined commands
func (p *Project) afterBefore(command string) (errors string, logs string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command = strings.Replace(strings.Replace(command, "'", "", -1), "\"", "", -1)
	c := strings.Split(command, " ")
	build := exec.Command(c[0], c[1:]...)
	build.Dir = p.base
	build.Stdout = &stdout
	build.Stderr = &stderr
	err := build.Run()
	// check if log
	logs = stdout.String()
	if err != nil {
		errors = stderr.String()
		return errors, logs
	}
	return "", logs
}
