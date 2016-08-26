package realize

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

// The Project struct defines the informations about a project
type Project struct {
	reload  time.Time
	base    string
	Name    string       `yaml:"app_name,omitempty"`
	Path    string       `yaml:"app_path,omitempty"`
	Run     bool         `yaml:"app_run,omitempty"`
	Bin     bool         `yaml:"app_bin,omitempty"`
	Build   bool         `yaml:"app_build,omitempty"`
	Fmt     bool         `yaml:"app_fmt,omitempty"`
	Params  []string		 `yaml:"app_params,omitempty"`
	Watcher Watcher      `yaml:"app_watcher,omitempty"`
}

// GoRun  is an implementation of the bin execution
func (p *Project) GoRun(channel chan bool, runner chan bool, wr *sync.WaitGroup) error {

	stop := make(chan bool, 1)
	var build *exec.Cmd
	if len(p.Params) != 0 {
		build = exec.Command(filepath.Join(os.Getenv("GOBIN"), filepath.Base(p.Path)), p.Params...)
	}	else{
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
				log.Println(pname(p.Name, 3), ":", BlueS(in.Text()))
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
func (p *Project) GoBuild() (error, string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	build := exec.Command("go", "build")
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return err, stderr.String()
	}
	return nil, ""
}

// GoInstall is an implementation of the "go install"
func (p *Project) GoInstall() (error, string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	err := os.Setenv("GOBIN", filepath.Join(os.Getenv("GOPATH"), "bin"))
	if err != nil {
		return nil, ""
	}
	build := exec.Command("go", "install")
	build.Dir = p.base
	build.Stdout = &out
	build.Stderr = &stderr
	if err := build.Run(); err != nil {
		return err, stderr.String()
	}
	return nil, ""
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
