package realize

import (
	"time"
	"os/exec"
	"os"
	"bytes"
	"bufio"
	"log"
)

type Project struct {
	reload  time.Time
	Name    string `yaml:"app_name,omitempty"`
	Path    string `yaml:"app_path,omitempty"`
	Main    string `yaml:"app_main,omitempty"`
	Run     bool `yaml:"app_run,omitempty"`
	Bin     bool `yaml:"app_bin,omitempty"`
	Build   bool `yaml:"app_build,omitempty"`
	Watcher Watcher `yaml:"app_watcher,omitempty"`
}

func (p *Project) GoRun(channel chan bool) error {
	base, _ := os.Getwd()
	build := exec.Command("go", "run", p.Main)
	path := base + p.Path
	build.Dir = path

	stdout, err := build.StdoutPipe()
	if err != nil {
		Fail(err.Error())
	}
	if err := build.Start(); err != nil {
		Fail(err.Error())
	}

	in := bufio.NewScanner(stdout)
	for in.Scan() {
		select {
		default:
			log.Println(p.Name + ":", in.Text())
		case <-channel:
			return nil
		}
	}
	return nil
}

func (p *Project) GoBuild() error {
	var out bytes.Buffer
	base, _ := os.Getwd()
	path := base + p.Path

	// create bin dir
	if _, err := os.Stat(path + "bin"); err != nil {
		if err = os.Mkdir(path + "bin", 0777); err != nil {
			return err
		}
	}

	build := exec.Command("go", "build", path + p.Main)
	build.Dir = path + "bin"
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

