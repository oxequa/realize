package realize

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"log"
	"strings"
	"time"
	"sync"
)

type Watcher struct {
	// different before and after on re-run?
	Before  []string `yaml:"before,omitempty"`
	After   []string `yaml:"after,omitempty"`
	Paths   []string `yaml:"paths,omitempty"`
	Ignore  []string `yaml:"ignore_paths,omitempty"`
	Exts    []string `yaml:"exts,omitempty"`
	Preview bool `yaml:"preview,omitempty"`
}

func (h *Config) Watch() error {
	if err := h.Read(); err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			_, h.Projects[k].Path = slash(h.Projects[k].Path)
			go h.Projects[k].Watching()
		}
		wg.Wait()
		return nil
	} else {
		return err
	}
}

func (p *Project) Watching() {

	channel := make(chan bool,1)
	var wr sync.WaitGroup
	var watcher *fsnotify.Watcher
	watcher, _ = fsnotify.NewWatcher()
	defer func() {
		watcher.Close()
		wg.Done()
	}()

	walk := func(path string, info os.FileInfo, err error) error {
		if !ignore(path, p.Watcher.Ignore) {
			if (info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.Contains(path, "/.")) || (inArray(filepath.Ext(path), p.Watcher.Exts)) {
				if p.Watcher.Preview {
					fmt.Println(p.Name + ": \t" + path)
				}
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
			}
		}
		return nil
	}
	routines := func(){
		channel = make(chan bool)
		wr.Add(1)
		go p.build(); p.install(); p.run(channel, &wr);
	}

	for _, dir := range p.Watcher.Paths {
		base, _ := os.Getwd()
		// check main existence
		if _, err := os.Stat(base + p.Path + dir + p.Main); err != nil {
			Fail(p.Name + ": \t" + base + p.Path + dir + p.Main + " doesn't exist. Main is required")
			return
		}

		base = base + p.Path
		if check, _ := slash(dir); check != true && len(dir) >= 1 {
			base = base + p.Path + dir
		}
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, walk); err != nil {
				Fail(err.Error())
			}
			if check, _ := slash(dir); check == true && len(dir) <= 1  {
				break
			}
		} else {
			Fail(p.Name + ": \t" + base + " path doesn't exist")
		}
	}

	routines()

	fmt.Println(red("\n Watching: '" + p.Name + "'\n"))

	p.reload = time.Now().Truncate(time.Second)

	for {
		select {
		case event := <-watcher.Events:
			if time.Now().Truncate(time.Second).After(p.reload) {
				if event.Op & fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				if _, err := os.Stat(event.Name); err == nil {
					i := strings.Index(event.Name, filepath.Ext(event.Name))
					log.Println(green(p.Name + ":"), event.Name[:i])

					// stop and run again
					close(channel)
					wr.Wait()
					routines()

					p.reload = time.Now().Truncate(time.Second)
				}
			}
		case err := <-watcher.Errors:
			Fail(err.Error())
		}
	}
}

func (p *Project) install() {
	if p.Bin {
		LogSuccess(p.Name + ": Installing..")
		if err := p.GoInstall(); err != nil {
			Fail(err.Error())
			return
		} else {
			LogSuccess(p.Name + ": Installed")
			return
		}
	}
	return
}

func (p *Project) build() {
	if p.Build {
		LogSuccess(p.Name + ": Building..")
		if err := p.GoBuild(); err != nil {
			Fail(err.Error())
			return
		} else {
			LogSuccess(p.Name + ": Builded")
			return
		}
	}
	return
}

func (p *Project) run(channel chan bool,  wr *sync.WaitGroup) {
	if p.Run {
		LogSuccess(p.Name + ": Running..")
		go p.GoRun(channel, wr)
		LogSuccess(p.Name + ": Runned")
	}
	return
}

func inArray(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func ignore(str string, list []string) bool {
	base, _ := os.Getwd()
	for _, v := range list {
		_, v = slash(v)
		if strings.Contains(str, base + v) {
			return true
		}
	}
	return false
}

func slash(str string) (bool, string){
	if string(str[0]) == "/" {
		return true, str
	}
	return false, "/"+str
}