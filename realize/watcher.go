package realize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/urfave/cli.v2"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

// Fast method run a project from his working directory without makes a config file
func (h *Config) Fast(params *cli.Context) error {
	fast := h.Projects[0]
	// Takes the values from config if wd path match someone else
	if params.Bool("config") {
		if err := h.Read(); err == nil {
			for _, val := range h.Projects {
				if fast.Path == val.Path {
					fast = val
				}
			}
		}
	}
	wg.Add(1)
	fast.Path = slash(fast.Path)
	go fast.Watching()
	wg.Wait()
	return nil
}

// Watching method is the main core. It manages the livereload and the watching
func (p *Project) Watching() {

	var wr sync.WaitGroup
	var watcher *fsnotify.Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(strings.ToUpper(pname(p.Name, 1)), ": \t", Red(err.Error()))
	}
	channel := make(chan bool, 1)
	base, err := os.Getwd()
	if err != nil {
		log.Println(pname(p.Name, 1), ": \t", Red(err.Error()))
	}

	walk := func(path string, info os.FileInfo, err error) error {
		if !p.ignore(path) {
			if (info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.Contains(path, "/.")) || (inArray(filepath.Ext(path), p.Watcher.Exts)) {
				if p.Watcher.Preview {
					fmt.Println(pname(p.Name, 1) + ": \t" + path)
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
		err = p.fmt()
		if err == nil {
			wr.Add(1)
			go p.build()
			go p.install(channel, &wr)
		}
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
			fmt.Println(pname(p.Name, 1), ":\t", Red(base+" path doesn't exist"))
		}
	}
	fmt.Println(Red("Watching: " + pname(p.Name, 1) + "\n"))
	routines()
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
						log.Println(pname(p.Name, 4), ":", Magenta(event.Name[:i]))

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
func (p *Project) install(channel chan bool, wr *sync.WaitGroup) {
	if p.Bin {
		log.Println(pname(p.Name, 1), ":", "Installing..")
		start := time.Now()
		if err := p.GoInstall(); err != nil {
			log.Println(pname(p.Name, 1), ":", Red(err.Error()))
			wr.Done()
		} else {
			log.Println(pname(p.Name, 5), ":", Green("Installed")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), "s"))
			if p.Run {
				runner := make(chan bool, 1)
				log.Println(pname(p.Name, 1), ":", "Running..")
				start = time.Now()
				go p.GoRun(channel, runner, wr)
				for {
					select {
					case <-runner:
						log.Println(pname(p.Name, 5), ":", Green("Has been run")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), "s"))
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
		log.Println(pname(p.Name, 1), ":", "Building..")
		start := time.Now()
		if err := p.GoBuild(); err != nil {
			log.Println(pname(p.Name, 1), ":", Red(err.Error()))
		} else {
			log.Println(pname(p.Name, 5), ":", Green("Builded")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), "s"))
		}
		return
	}
	return
}

// Build call an implementation of the "gofmt"
func (p *Project) fmt() error {
	if p.Fmt {
		if msg, err := p.GoFmt(); err != nil {
			log.Println(pname(p.Name, 1), Red("There are some errors:"))
			fmt.Println(msg)
			return err
		}
	}
	return nil
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
	if len(str) == 0 || string(str[0]) != "/" {
		str = "/" + str
	}
	if string(str[len(str)-1]) == "/" {
		if string(str) == "/" {
			return str
		} else {
			str = str[0 : len(str)-1]
		}
	}
	return str
}

// defines the colors scheme for the project name
func pname(name string, color int) string {
	switch color {
	case 1:
		name = Yellow("[") + strings.ToUpper(name) + Yellow("]")
		break
	case 2:
		name = Yellow("[") + Red(strings.ToUpper(name)) + Yellow("]")
		break
	case 3:
		name = Yellow("[") + Blue(strings.ToUpper(name)) + Yellow("]")
		break
	case 4:
		name = Yellow("[") + Magenta(strings.ToUpper(name)) + Yellow("]")
		break
	case 5:
		name = Yellow("[") + Green(strings.ToUpper(name)) + Yellow("]")
		break
	}
	return name
}
