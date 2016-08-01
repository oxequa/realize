package realize

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
	"path/filepath"
	"os"
	"log"
	"strings"
	"sync"
	"time"
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

	var wg sync.WaitGroup

	var watcher *fsnotify.Watcher

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

	watch := func(val Project){

		watcher, _ = fsnotify.NewWatcher()

		// run, bin, build
		val.reload = time.Now().Truncate(time.Second)

		for _, dir := range val.Watcher.Paths {
			path, _ := os.Getwd()
			current = val.Watcher
			// add dir of project 
			if err := filepath.Walk(path + dir, walk); err != nil {
				fmt.Println(err)
			}
		}
		for {
			select {
				case event := <-watcher.Events:
					if time.Now().Truncate(time.Second).After(val.reload) {
						if event.Op & fsnotify.Chmod == fsnotify.Chmod {
							continue
						}
						if _, err := os.Stat(event.Name); err == nil {
							log.Println("event:", event)
							// run, bin, build
							val.reload = time.Now().Truncate(time.Second)
						}
					}
				case err := <-watcher.Errors:
					log.Println("error:", err)
			}
		}
		watcher.Close()
		wg.Done()
	}

	// add to watcher
	if err := h.Read(); err == nil {

		// loop projects
		wg.Add(len(h.Projects))
		for _, val := range h.Projects {
			go watch(val)
		}
		wg.Wait()

		return nil
	}else{
		return err
	}
}