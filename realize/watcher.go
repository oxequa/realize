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

func (p *Project) Watching(){

	var watcher *fsnotify.Watcher
	watcher, _ = fsnotify.NewWatcher()
	defer func(){
		watcher.Close()
		wg.Done()
	}()

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

	// run, bin, build
	p.GoBuild()

	p.reload = time.Now().Truncate(time.Second)

	for _, dir := range p.Watcher.Paths {

		base, _ := os.Getwd()

		// check path existence
		if _, err := os.Stat(base + p.Path + dir); err == nil {
			if err := filepath.Walk(base + p.Path + dir, walk); err != nil {
				fmt.Println(err)
			}
		}else{
			fmt.Println(red(p.Name + ": \t"+base + p.Path + dir +" path doesn't exist"))
		}
	}

	fmt.Println(red("Watching: '"+ p.Name +"'\n"))

	for {
		select {
		case event := <-watcher.Events:
			if time.Now().Truncate(time.Second).After(p.reload) {
				if event.Op & fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				if _, err := os.Stat(event.Name); err == nil {
					i := strings.Index(event.Name, filepath.Ext(event.Name))
					log.Println(green(p.Name+":")+"\t", event.Name[:i])
					// run, bin, build
					p.reload = time.Now().Truncate(time.Second)
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (h *Config) Watch() error{
	if err := h.Read(); err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			go h.Projects[k].Watching()
		}
		wg.Wait()
		return nil
	}else{
		return err
	}
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