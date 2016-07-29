package realize

import (
	"github.com/fsnotify/fsnotify"
	"path/filepath"
	"os"
)

func (h *Config) Watch() error{

	watcher, err := fsnotify.NewWatcher()
	if err != nil{
		panic(err)
	}

	walking := func(path string, info os.FileInfo, err error) error{
		if info.IsDir() {
			if err = watcher.Add(path); err != nil {
				return filepath.SkipDir
			}
		}
		return nil
	}

	// check file
	if err := h.Read(); err == nil {
		// loop projects
		for _, val := range h.Projects {
			// add paths to watcher
			for _, path := range val.Watcher.Paths {
				if err := filepath.Walk(path, walking); err != nil{
					panic(err)
				}
			}
		}
		return nil
	}else{
		return err
	}
}