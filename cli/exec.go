package cli

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// GoRun  is an implementation of the bin execution
func (p *Project) GoRun(channel chan bool, runner chan bool, wr *sync.WaitGroup) error {

	stop := make(chan bool, 1)
	var build *exec.Cmd
	if len(p.Params) != 0 {
		build = exec.Command(filepath.Join(os.Getenv("GOBIN"), filepath.Base(p.Path)), p.Params...)
	} else {
		build = exec.Command(filepath.Join(os.Getenv("GOBIN"), filepath.Base(p.Path)))
	}
	build.Dir = p.base
	defer func() {
		if err := build.Process.Kill(); err != nil {
			log.Fatal(Red("Failed to stop: "), Red(err))
		}
		log.Println(pname(p.Name, 2), ":", RedS("Stopped"))
		wr.Done()
	}()

	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()

	// Read stdout and stderr in same var
	outputs := io.MultiReader(stdout, stderr)
	if err != nil {
		log.Println(Red(err.Error()))
		return err
	}
	if err := build.Start(); err != nil {
		log.Println(Red(err.Error()))
		return err
	}
	close(runner)

	in := bufio.NewScanner(outputs)
	go func() {
		for in.Scan() {
			select {
			default:
				if p.Watcher.Output["cli"] {
					log.Println(pname(p.Name, 3), ":", BlueS(in.Text()))
				}
				if p.Watcher.Output["file"] {
					path := filepath.Join(p.base, App.Blueprint.Files["output"])
					f := create(path)
					t := time.Now()
					if _, err := f.WriteString(t.Format("2006-01-02 15:04:05") + " : " + in.Text() + "\r\n"); err != nil {
						log.Fatal(err)
					}
				}
			}
		}
		close(stop)
	}()

	for {
		select {
		case <-channel:
			return nil
		case <-stop:
			return nil
		}
	}
}

// GoBuild is an implementation of the "go build"
func (p *Project) GoBuild() (string, error) {
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
func (p *Project) GoInstall() (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	err := os.Setenv("GOBIN", filepath.Join(os.Getenv("GOPATH"), "bin"))
	if err != nil {
		return "", nil
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
func (p *Project) GoFmt(path string) (io.Writer, error) {
	var out bytes.Buffer
	build := exec.Command("gofmt", "-s", "-w", "-e", path)
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &out
	if err := build.Run(); err != nil {
		return build.Stderr, err
	}
	return nil, nil
}

// GoTest is an implementation of the go test
func (p *Project) GoTest(path string) (io.Writer, error) {
	var out bytes.Buffer
	build := exec.Command("go", "test")
	build.Dir = path
	build.Stdout = &out
	build.Stderr = &out
	if err := build.Run(); err != nil {
		return build.Stdout, err
	}
	return nil, nil
}
