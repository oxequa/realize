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
	"bytes"
)

var wg sync.WaitGroup

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

func Watching(val Project){

	var current Project
	var watcher *fsnotify.Watcher
	watcher, _ = fsnotify.NewWatcher()
	defer func(){
		watcher.Close()
		wg.Done()
	}()

	walk := func(path string, info os.FileInfo, err error) error{
		if !Ignore(path,current.Watcher.Ignore) {
			if info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.Contains(path, "/.") {
				fmt.Println(current.Name +": "+path)
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
			} else if InArray(filepath.Ext(path), current.Watcher.Exts) {
				fmt.Println(current.Name +": "+path)
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
			}
		}
		return nil
	}

	// run, bin, build

	val.reload = time.Now().Truncate(time.Second)

	for _, dir := range val.Watcher.Paths {

		var base bytes.Buffer
		path, _ := os.Getwd()
		current = val
		split := strings.Split(val.Main, "/")

		// get base path from mail field
		for key, str := range split{
			if(key < len(split)-1) {
				base.WriteString("/" + str)
			}
		}

		if err := filepath.Walk(path + base.String() + dir, walk); err != nil {
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

					i := strings.Index(event.Name, filepath.Ext(event.Name))
					log.Println("event:", event.Name[:i])

					// run, bin, build

					val.reload = time.Now().Truncate(time.Second)
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
		for _, val := range h.Projects {
			go Watching(val)
		}
		wg.Wait()
		return nil
	}else{
		return err
	}
}