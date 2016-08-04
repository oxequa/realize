package realize

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"log"
	"strings"
	"time"
)

type Watcher struct{
	// different before and after on re-run?
	Before []string `yaml:"before,omitempty"`
	After []string `yaml:"after,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
	Ignore []string `yaml:"ignore_paths,omitempty"`
	Exts []string `yaml:"exts,omitempty"`
	Preview bool `yaml:"preview,omitempty"`
}

func (h *Config) Watch() error{
	if err := h.Read(); err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			if string(h.Projects[k].Path[0]) != "/" {
				h.Projects[k].Path = "/"+h.Projects[k].Path
			}
			go h.Projects[k].Watching()
		}
		wg.Wait()
		return nil
	}else{
		return err
	}
}

func (p *Project) Watching(){

	var watcher *fsnotify.Watcher
	channel := make(chan bool)
	watcher, _ = fsnotify.NewWatcher()

	walk := func(path string, info os.FileInfo, err error) error{
		if !Ignore(path,p.Watcher.Ignore) {
			if (info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.Contains(path, "/.")) || (InArray(filepath.Ext(path), p.Watcher.Exts)){
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

	for _, dir := range p.Watcher.Paths {
		base, _ := os.Getwd()
		// check path existence
		if _, err := os.Stat(base + p.Path + dir); err == nil {
			if err := filepath.Walk(base + p.Path + dir, walk); err != nil {
				Fail(err.Error())
			}
		}else{
			Fail(p.Name + ": \t"+base + p.Path + dir +" path doesn't exist")
		}
	}

	// go build, install, run
	go p.build(); p.install(); p.run(channel);

	fmt.Println(red("\n Watching: '"+ p.Name +"'\n"))

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
						log.Println(green(p.Name+":"), event.Name[:i])

						// stop and run again
						close(channel)
						channel = make(chan bool)
						go p.build(); p.install(); p.run(channel);

						p.reload = time.Now().Truncate(time.Second)
					}
				}
			case err := <-watcher.Errors:
				Fail(err.Error())
		}
	}

	watcher.Close()
	wg.Done()
}

func (p *Project) install(){
	if p.Bin {
		LogSuccess(p.Name + ": Installing..")
		if err := p.GoInstall(); err != nil{
			Fail(err.Error())
			return
		}else{
			LogSuccess(p.Name + ": Installed")
			return
		}
	}
	return
}

func (p *Project) build(){
	if p.Build {
		LogSuccess(p.Name + ": Building..")
		if err := p.GoBuild(); err != nil{
			Fail(err.Error())
			return
		}else{
			LogSuccess(p.Name + ": Builded")
			return
		}
	}
	return
}

func (p *Project) run(channel chan bool){
	if p.Run {
		LogSuccess(p.Name + ": Running..")
		go p.GoRun(channel)
		LogSuccess(p.Name + ": Runned")
	}
	return
}

func InArray(str string, list []string) bool{
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func Ignore(str string, list []string) bool{
	for _, v := range list {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}