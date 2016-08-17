package realize

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Project struct {
	reload  time.Time
	base    string
	Name    string  `yaml:"app_name,omitempty"`
	Path    string  `yaml:"app_path,omitempty"`
	Main    string  `yaml:"app_main,omitempty"`
	Run     bool    `yaml:"app_run,omitempty"`
	Bin     bool    `yaml:"app_bin,omitempty"`
	Build   bool    `yaml:"app_build,omitempty"`
	Watcher Watcher `yaml:"app_watcher,omitempty"`
}

func (p *Project) GoRun(channel chan bool, runner chan bool, wr *sync.WaitGroup) error {
	name := strings.Split(p.Path, "/")
	stop := make(chan bool, 1)
	var run string

	if len(name) == 1 {
		name := strings.Split(p.base, "/")
		run = name[len(name)-1]
	} else {
		run = name[len(name)-1]
	}
	build := exec.Command(os.Getenv("GOPATH") + slash("bin") + slash(run))
	build.Dir = p.base
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
	close(runner)

	in := bufio.NewScanner(stdout)
	go func() {
		for in.Scan() {
			select {
			default:
				log.Println(p.Name+":", in.Text())
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

	return nil
}

func (p *Project) GoBuild() error {
	var out bytes.Buffer

	// create bin dir
	if _, err := os.Stat(p.base + "/bin"); err != nil {
		if err = os.Mkdir(p.base+"/bin", 0777); err != nil {
			return err
		}
	}
	build := exec.Command("go", "build", p.base+p.Main)
	build.Dir = p.base + "/bin"
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
