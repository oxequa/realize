package watcher

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/tockins/realize/style"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GoRun  is an implementation of the bin execution
func (p *Project) goRun(channel chan bool, runner chan bool, wr *sync.WaitGroup) error {
	var build *exec.Cmd
	var args []string
	isErrorText := func(string) bool {
		return false
	}
	errRegexp, err := regexp.Compile(p.ErrorOutputPattern)
	if err != nil {
		msg := fmt.Sprintln(p.pname(p.Name, 3), ":", style.Blue.Regular(err.Error()))
		out := BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Run"}
		p.stamp("error", out, msg, "")
	} else {
		isErrorText = func(t string) bool {
			return errRegexp.MatchString(t)
		}
	}
	for _, arg := range p.Args {
		arr := strings.Fields(arg)
		args = append(args, arr...)
	}

	if _, err := os.Stat(filepath.Join(p.base, p.path)); err == nil {
		p.path = filepath.Join(p.base, p.path)
	}
	if _, err := os.Stat(filepath.Join(p.base, p.path+".exe")); err == nil {
		p.path = filepath.Join(p.base, p.path+".exe")
	}

	if _, err := os.Stat(filepath.Join(getEnvPath("GOBIN"), filepath.Base(p.path))); err == nil {
		build = exec.Command(filepath.Join(getEnvPath("GOBIN"), filepath.Base(p.path)), args...)
	} else if _, err := os.Stat(filepath.Join(getEnvPath("GOBIN"), filepath.Base(p.path)) + ".exe"); err == nil {
		build = exec.Command(filepath.Join(getEnvPath("GOBIN"), filepath.Base(p.path))+".exe", args...)
	} else {
		p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Can't run a not compiled project"})
		p.Fatal(err, "Can't run a not compiled project", ":")
	}

	defer func() {
		if err := build.Process.Kill(); err != nil {
			p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Failed to stop: " + err.Error()})
			p.Fatal(err, "Failed to stop", ":")
		}
		msg := fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Regular("Ended"))
		out := BufferOut{Time: time.Now(), Text: "Ended", Type: "Go Run"}
		p.stamp("log", out, msg, "")
		wr.Done()
	}()

	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()
	if err != nil {
		log.Println(style.Red.Bold(err.Error()))
		return err
	}
	if err := build.Start(); err != nil {
		log.Println(style.Red.Bold(err.Error()))
		return err
	}
	close(runner)

	execOutput, execError := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOutput, stopError := make(chan bool, 1), make(chan bool, 1)
	scanner := func(stop chan bool, output *bufio.Scanner, isError bool) {
		for output.Scan() {
			text := output.Text()
			msg := fmt.Sprintln(p.pname(p.Name, 3), ":", style.Blue.Regular(text))
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
	args := []string{"build"}
	args = arguments(args, p.Cmds.Build.Args)
	build := exec.Command("go", args...)
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
	err := os.Setenv("GOBIN", filepath.Join(getEnvPath("GOPATH"), "bin"))
	if err != nil {
		return "", err
	}
	args := []string{"install"}
	for _, arg := range p.Cmds.Bin.Args {
		arr := strings.Fields(arg)
		args = append(args, arr...)
	}
	build := exec.Command("go", args...)
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
	if s := filepath.Ext(dir); s != "" && s != ".go" {
		return "", nil
	}
	var out, stderr bytes.Buffer
	build := exec.Command(name, cmd...)
	build.Dir = dir
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String() + out.String(), err
	}
	return "", nil
}

// Exec an additional command from a defined path if specified
func (p *Project) command(cmd Command) (errors string, logs string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := strings.Replace(strings.Replace(cmd.Command, "'", "", -1), "\"", "", -1)
	c := strings.Split(command, " ")
	build := exec.Command(c[0], c[1:]...)
	build.Dir = p.base
	if cmd.Path != "" {
		if strings.Contains(cmd.Path, p.base) {
			build.Dir = cmd.Path
		} else {
			build.Dir = filepath.Join(p.base, cmd.Path)
		}
	}
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
