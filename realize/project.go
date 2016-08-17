package realize

import (
	"time"
	"os/exec"
	"os"
	"bytes"
	"bufio"
	"log"
	"sync"
	"strings"
)

type Project struct {
	reload  time.Time
	Base 	string
	Name    string `yaml:"app_name,omitempty"`
	Path    string `yaml:"app_path,omitempty"`
	Main    string `yaml:"app_main,omitempty"`
	Run     bool `yaml:"app_run,omitempty"`
	Bin     bool `yaml:"app_bin,omitempty"`
	Build   bool `yaml:"app_build,omitempty"`
	Watcher Watcher `yaml:"app_watcher,omitempty"`
}

func (p *Project) GoRun(channel chan bool, wr *sync.WaitGroup) error {
	name := strings.Split(p.Path, "/")
	build := exec.Command(name[len(name)-1], os.Getenv("PATH"))
	build.Dir = p.Base
	defer func() {
		if err := build.Process.Kill(); err != nil {
			log.Fatal("failed to stop: ", err)
		}
		LogFail(p.Name + ": Stopped")
		wr.Done()
	}()

	stdout, err := build.StdoutPipe()
	if err != nil {
		Fail(err.Error())
	}
	if err := build.Start(); err != nil {
		Fail(err.Error())
	}

	in := bufio.NewScanner(stdout)
	go func() {
		for in.Scan() {
			select {
			default:
				log.Println(p.Name + ":", in.Text())
			}
		}
	}()

	for{
		select {
		case <-channel:
			return nil
		}
	}

	return nil
}

func (p *Project) GoBuild() error {
	var out bytes.Buffer

	// create bin dir
	if _, err := os.Stat(p.Base + "/bin"); err != nil {
		if err = os.Mkdir(p.Base + "/bin", 0777); err != nil {
			return err
		}
	}
	build := exec.Command("go", "build", p.Base + p.Main)
	build.Dir = p.Base + "/bin"
	build.Stdout = &out
	if err := build.Run(); err != nil {
		return err
	}
	return nil
}

func (p *Project) GoInstall() error {
	var out bytes.Buffer
	base, _ := os.Getwd()
	path := base + p.Path

	build := exec.Command("go", "install")
	build.Dir = path
	build.Stdout = &out
	if err := build.Run(); err != nil {
		return err
	}
	return nil
}

