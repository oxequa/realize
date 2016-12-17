package cli

import (
	"bufio"
	"bytes"
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
		p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Ended"})
		log.Println(p.pname(p.Name, 2), ":", p.Red.Regular("Ended"))
		p.sync()
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
				if isError {
					p.Buffer.StdErr = append(p.Buffer.StdErr, BufferOut{Time: time.Now(), Text: output.Text(), Type: "Go Run"})
				} else {
					p.Buffer.StdOut = append(p.Buffer.StdOut, BufferOut{Time: time.Now(), Text: output.Text()})
				}
				p.sync()
				if p.Cli.Streams {
					log.Println(p.pname(p.Name, 3), ":", p.Blue.Regular(output.Text()))
				}
				if p.File.Streams {
					f := p.Create(p.base, p.parent.Resources.Output)
					t := time.Now()
					if _, err := f.WriteString(t.Format("2006-01-02 15:04:05") + " : " + output.Text() + "\r\n"); err != nil {
						p.Fatal(err, "")
					}
				}
			}
		}
		close(stop)
	}
	p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: "Started"})
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
	defer func() {
		p.sync()
	}()
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
	defer func() {
		p.sync()
	}()
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

// GoFmt is an implementation of the gofmt
func (p *Project) goFmt(path string) (string, error) {
	var out, stderr bytes.Buffer
	build := exec.Command("gofmt", "-s", "-w", "-e", path)
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// GoTest is an implementation of the go test
func (p *Project) goTest(path string) (string, error) {
	var out, stderr bytes.Buffer
	build := exec.Command("go", "test")
	build.Dir = path
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// GoGenerate is an implementation of the go test
func (p *Project) goGenerate(path string) (string, error) {
	var out, stderr bytes.Buffer
	build := exec.Command("go", "generate")
	build.Dir = path
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return stderr.String(), err
	}
	return "", nil
}

// Cmds exec a list of defined commands
func (p *Project) cmds(cmds []string) (errors []string) {
	defer func() {
		p.sync()
	}()
	for _, cmd := range cmds {
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd := strings.Replace(strings.Replace(cmd, "'", "", -1), "\"", "", -1)
		c := strings.Split(cmd, " ")
		build := exec.Command(c[0], c[1:]...)
		build.Dir = p.base
		build.Stdout = &out
		build.Stderr = &stderr
		if err := build.Run(); err != nil {
			errors = append(errors, stderr.String())
			return errors
		}
	}
	return nil
}

// Sync datas with the web server
func (p *Project) sync() {
	go func() {
		p.parent.Sync <- "sync"
	}()
}
