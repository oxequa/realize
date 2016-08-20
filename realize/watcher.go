package realize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"log"
)

// The Watcher struct defines the livereload's logic
type Watcher struct {
	// different before and after on re-run?
	Before  []string `yaml:"before,omitempty"`
	After   []string `yaml:"after,omitempty"`
	Paths   []string `yaml:"paths,omitempty"`
	Ignore  []string `yaml:"ignore_paths,omitempty"`
	Exts    []string `yaml:"exts,omitempty"`
	Preview bool     `yaml:"preview,omitempty"`
}

// Watch method adds the given paths on the Watcher
func (h *Config) Watch() error {
	err := h.Read()
	if err == nil {
		// loop projects
		wg.Add(len(h.Projects))
		for k := range h.Projects {
			h.Projects[k].Path = slash(h.Projects[k].Path)
			go h.Projects[k].Watching()
		}
		wg.Wait()
		return nil
	}
	return err
}

// Watching method is the main core. It manages the livereload and the watching
func (p *Project) Watching() {

	var wr sync.WaitGroup
	var watcher *fsnotify.Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(Redl(p.Name),": \t", Red(err.Error()))
	}
	channel := make(chan bool, 1)
	base, err := os.Getwd()
	if err != nil {
		log.Println(Redl(p.Name),": \t", Red(err.Error()))
	}

	walk := func(path string, info os.FileInfo, err error) error {
		if !p.ignore(path) {
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
	routines := func() {
		channel = make(chan bool)
		wr.Add(1)
		go p.build()
		go p.install(channel,&wr)
	}
	end := func() {
		watcher.Close()
		wg.Done()
	}
	defer end()

	p.base = base + p.Path

	for _, dir := range p.Watcher.Paths {
		// check main existence
		dir = slash(dir)

		base = p.base + dir
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, walk); err != nil {
				log.Println(Red(err.Error()))
			}
		} else {
			fmt.Println(Redl(p.Name), ":\t", Red(base + " path doesn't exist"))
		}
	}
	routines()
	fmt.Println(Red("Watching: '" + p.Name + "'\n"))
	p.reload = time.Now().Truncate(time.Second)
	for {
		select {
		case event := <-watcher.Events:
			if time.Now().Truncate(time.Second).After(p.reload) {
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				if _, err := os.Stat(event.Name); err == nil {
					i := strings.Index(event.Name, filepath.Ext(event.Name))
					if event.Name[:i] != "" {
						log.Println(Magenta(p.Name),":", Magenta("File changed"), "->", Blue(event.Name[:i]))

						// stop and run again
						close(channel)
						wr.Wait()
						routines()

						p.reload = time.Now().Truncate(time.Second)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println(Red(err.Error()))
		}
	}
}

// Install call an implementation of the "go install"
func (p *Project) install(channel chan bool,wr *sync.WaitGroup) {
	if p.Bin {
		log.Println(Greenl(p.Name), ":", Greenl("Installing.."))
		start := time.Now()
		if err := p.GoInstall(); err != nil {
			log.Println(Redl(p.Name),":",Red(err.Error()))
			wr.Done()
		} else {
			log.Println(Greenl(p.Name),":", Green("Installed")+ " after",  Magenta(time.Since(start)))
			if p.Run {
				runner := make(chan bool, 1)
				log.Println(Greenl(p.Name), ":", Greenl("Running.."))
				start = time.Now()
				go p.GoRun(channel, runner, wr)
				for {
					select {
					case <-runner:
						log.Println(Greenl(p.Name), ":", Green("Has been run") + " after", Magenta(time.Since(start)))
						return
					}
				}
			}
		}
	}
	return
}

// Build call an implementation of the "go build"
func (p *Project) build() {
	if p.Build {
		log.Println(Greenl(p.Name), ":", Greenl("Building.."))
		start := time.Now()
		if err := p.GoBuild(); err != nil {
			log.Println(Redl(p.Name), ":", Red(err.Error()))
		} else {
			log.Println(Greenl(p.Name),":", Green("Builded")+ " after",  Magenta(time.Since(start)))
		}
		return
	}
	return
}

// Ignore validates a path
func (p *Project) ignore(str string) bool {
	for _, v := range p.Watcher.Ignore {
		v = slash(v)
		if strings.Contains(str, p.base+v) {
			return true
		}
	}
	return false
}

// check if a string is inArray
func inArray(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// add a slash at the beginning if not exist
func slash(str string) string {
	if string(str[0]) != "/" {
		str = "/" + str
	}
	if string(str[len(str)-1]) == "/" {
		if string(str) == "/" {
			str = ""
		} else {
			str = str[0 : len(str)-2]
		}
	}
	return str
}
