package cli

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Watching method is the main core. It manages the livereload and the watching
func (p *Project) watching() {
	var wr sync.WaitGroup
	var watcher *fsnotify.Watcher
	channel, exit := make(chan bool, 1), make(chan bool, 1)
	p.path = p.Path
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(strings.ToUpper(p.pname(p.Name, 1)), ":", p.Red.Bold(err.Error()))
		return
	}
	defer func() {
		watcher.Close()
		wg.Done()
	}()

	p.cmd(exit)
	if p.walks(watcher) != nil {
		log.Println(strings.ToUpper(p.pname(p.Name, 1)), ":", p.Red.Bold(err.Error()))
		return
	}
	go p.routines(channel, &wr)
	p.LastChangedOn = time.Now().Truncate(time.Second)
	// waiting for an event
	for {
		select {
		case event := <-watcher.Events:
			if time.Now().Truncate(time.Second).After(p.LastChangedOn) {
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					continue
				}
				if _, err := os.Stat(event.Name); err == nil {
					var ext string
					if index := strings.Index(filepath.Ext(event.Name), "_"); index == -1 {
						ext = filepath.Ext(event.Name)
					} else {
						ext = filepath.Ext(event.Name)[0:index]
					}
					i := strings.Index(event.Name, filepath.Ext(event.Name))
					file := event.Name[:i] + ext
					path := filepath.Dir(event.Name[:i])
					if event.Name[:i] != "" && inArray(ext, p.Watcher.Exts) {
						p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + file})
						go func() {
							p.parent.Sync <- "sync"
						}()
						fmt.Println(p.pname(p.Name, 4), p.Magenta.Bold(strings.ToUpper(ext[1:])+" changed"), p.Magenta.Bold(file))
						// stop and run again
						if p.Run {
							close(channel)
							channel = make(chan bool)
						}
						// handle multiple errors, need a better way
						p.fmt(file)
						p.test(path)
						p.generate(path)
						go p.routines(channel, &wr)
						p.LastChangedOn = time.Now().Truncate(time.Second)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println(p.Red.Bold(err.Error()))
		case <-exit:
			return
		}
	}
}

// Install calls an implementation of the "go install"
func (p *Project) install(channel chan bool, wr *sync.WaitGroup) {
	if p.Bin {
		log.Println(p.pname(p.Name, 1), ":", "Installing..")
		start := time.Now()
		if std, err := p.goInstall(); err != nil {
			log.Println(p.pname(p.Name, 1), ":", fmt.Sprint(p.Red.Bold(err)), std)
			wr.Done()
		} else {
			log.Println(p.pname(p.Name, 5), ":", p.Green.Regular("Installed")+" after", p.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			if p.Run {
				runner := make(chan bool, 1)
				log.Println(p.pname(p.Name, 1), ":", "Running..")
				start = time.Now()
				go p.goRun(channel, runner, wr)
				for {
					select {
					case <-runner:
						log.Println(p.pname(p.Name, 5), ":", p.Green.Regular("Has been run")+" after", p.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
						return
					}
				}
			}
		}
	}
	return
}

// Build calls an implementation of the "go build"
func (p *Project) build() {
	if p.Build {
		log.Println(p.pname(p.Name, 1), ":", "Building..")
		start := time.Now()
		if std, err := p.goBuild(); err != nil {
			log.Println(p.pname(p.Name, 1), ":", fmt.Sprint(p.Red.Bold(err)), std)
		} else {
			log.Println(p.pname(p.Name, 5), ":", p.Green.Regular("Builded")+" after", p.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
		}
	}
	return
}

// Fmt calls an implementation of the "go fmt"
func (p *Project) fmt(path string) error {
	if p.Fmt {
		if stream, err := p.goFmt(path); err != nil {
			log.Println(p.pname(p.Name, 1), p.Red.Bold("go Fmt"), p.Red.Bold("there are some errors in"), ":", p.Magenta.Bold(path))
			fmt.Println(stream)
			return err
		}
	}
	return nil
}

// Generate calls an implementation of the "go generate"
func (p *Project) generate(path string) error {
	if p.Generate {
		if stream, err := p.goGenerate(path); err != nil {
			log.Println(p.pname(p.Name, 1), p.Red.Bold("go Generate"), p.Red.Bold("there are some errors in"), ":", p.Magenta.Bold(path))
			fmt.Println(stream)
			return err
		}
	}
	return nil
}

// Cmd calls an wrapper for execute the commands after/before
func (p *Project) cmd(exit chan bool) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	cast := func(commands []string) {
		if errs := p.cmds(commands); errs != nil {
			for _, err := range errs {
				log.Println(p.pname(p.Name, 2), p.Red.Bold(err))
			}
		}
	}

	if len(p.Watcher.Before) > 0 {
		cast(p.Watcher.Before)
	}

	go func() {
		for {
			select {
			case <-c:
				if len(p.Watcher.After) > 0 {
					cast(p.Watcher.After)
				}
				close(exit)
			}
		}
	}()
}

// Test calls an implementation of the "go test"
func (p *Project) test(path string) error {
	if p.Test {
		if stream, err := p.goTest(path); err != nil {
			log.Println(p.pname(p.Name, 1), p.Red.Bold("go Test fails in "), ":", p.Magenta.Bold(path))
			fmt.Println(stream)
			return err
		}
	}
	return nil
}

// Walks the file tree of a project
func (p *Project) walks(watcher *fsnotify.Watcher) error {
	var files, folders int64
	wd, _ := os.Getwd()
	walk := func(path string, info os.FileInfo, err error) error {
		if !p.ignore(path) {
			if (info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.HasPrefix(path, ".")) && !strings.Contains(path, "/.") || (inArray(filepath.Ext(path), p.Watcher.Exts)) {
				if p.Watcher.Preview {
					fmt.Println(p.pname(p.Name, 1), ":", path)
				}
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
				if inArray(filepath.Ext(path), p.Watcher.Exts) {
					files++
					p.fmt(path)
				} else {
					folders++
					p.generate(path)
					p.test(path)
				}
			}
		}
		return nil
	}

	if p.path == "." || p.path == "/" {
		p.base = wd
		p.path = p.Wdir()
	} else if filepath.IsAbs(p.path) {
		p.base = p.path
	} else {
		p.base = filepath.Join(wd, p.path)
	}
	for _, dir := range p.Watcher.Paths {
		base := filepath.Join(p.base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, walk); err != nil {
				log.Println(p.Red.Bold(err.Error()))
			}
		} else {
			return errors.New(base + " path doesn't exist")
		}
	}
	fmt.Println(p.pname(p.Name, 1), p.Red.Bold("Watching"), p.Magenta.Bold(files), "file/s", p.Magenta.Bold(folders), "folder/s")
	return nil
}

// Ignore validates a path
func (p *Project) ignore(str string) bool {
	for _, v := range p.Watcher.Ignore {
		if strings.Contains(str, filepath.Join(p.base, v)) {
			return true
		}
	}
	return false
}

// Routines launches the following methods: run, build, install
func (p *Project) routines(channel chan bool, wr *sync.WaitGroup) {
	wr.Add(1)
	go p.build()
	go p.install(channel, wr)
	wr.Wait()
}

// Defines the colors scheme for the project name
func (p *Project) pname(name string, color int) string {
	switch color {
	case 1:
		name = p.Yellow.Regular("[") + strings.ToUpper(name) + p.Yellow.Regular("]")
		break
	case 2:
		name = p.Yellow.Regular("[") + p.Red.Bold(strings.ToUpper(name)) + p.Yellow.Regular("]")
		break
	case 3:
		name = p.Yellow.Regular("[") + p.Blue.Bold(strings.ToUpper(name)) + p.Yellow.Regular("]")
		break
	case 4:
		name = p.Yellow.Regular("[") + p.Magenta.Bold(strings.ToUpper(name)) + p.Yellow.Regular("]")
		break
	case 5:
		name = p.Yellow.Regular("[") + p.Green.Bold(strings.ToUpper(name)) + p.Yellow.Regular("]")
		break
	}
	return name
}
