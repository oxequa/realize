package realize

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"log"
	"strings"
)

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

func (h *Config) Watch() error{

	var current Watcher

	watcher, err := fsnotify.NewWatcher()
	if err != nil{
		fmt.Println(err)
	}

	walk := func(path string, info os.FileInfo, err error) error{
		if !Ignore(path,current.Ignore) {
			if info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.Contains(path, "/.") {
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
			} else if InArray(filepath.Ext(path), current.Exts) {
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
			}
		}
		return nil
	}

	watch := func(){
		for {
			select {
				case event := <-watcher.Events:
					log.Println("event:", event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
					}
				case err := <-watcher.Errors:
					log.Println("error:", err)
			}
		}
	}

	defer func(){
		watcher.Close()
		// kill process
	}()

	// add to watcher
	if err := h.Read(); err == nil {
		// loop projects
		for _, val := range h.Projects {
			// add paths
			for _, path := range val.Watcher.Paths {
				p, _ := os.Getwd()
				current = val.Watcher
				if err := filepath.Walk(p+path, walk); err != nil{
					fmt.Println(err)
				}
			}
		}

		// watch changes
		watch()

		// build

		// install

		// run

		return nil
	}else{
		return err
	}
}