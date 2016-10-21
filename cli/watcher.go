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

	sync := func() {
		p.parent.Sync <- "sync"
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(strings.ToUpper(pname(p.Name, 1)), ":", Red(err.Error()))
	}
	channel, exit := make(chan bool, 1), make(chan bool, 1)
	if err != nil {
		log.Println(pname(p.Name, 1), ":", Red(err.Error()))
	}
	end := func() {
		watcher.Close()
		wg.Done()
	}
	defer end()

	p.cmd(exit)
	err = p.walks(watcher)
	if err != nil {
		fmt.Println(pname(p.Name, 1), ":", Red(err.Error()))
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
						ext = filepath.Ext(event.Name)
						ext = ext[0:index]
					}

					i := strings.Index(event.Name, filepath.Ext(event.Name))
					if event.Name[:i] != "" && inArray(ext, p.Watcher.Exts) {
						p.Buffer.StdLog = append(p.Buffer.StdLog, BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + event.Name[:i] + ext})
						go sync()
						fmt.Println(pname(p.Name, 4), Magenta(strings.ToUpper(ext[1:])+" changed"), Magenta(event.Name[:i]+ext))
						// stop and run again
						if p.Run {
							close(channel)
							channel = make(chan bool)
						}

						err := p.fmt(event.Name[:i] + ext)
						if err != nil {
							p.Fatal("", err)
						} else {
							go p.routines(channel, &wr)
							p.LastChangedOn = time.Now().Truncate(time.Second)
						}
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println(Red(err.Error()))
		case <-exit:
			return
		}
	}
}

// Install calls an implementation of the "go install"
func (p *Project) install(channel chan bool, wr *sync.WaitGroup) {
	if p.Bin {
		log.Println(pname(p.Name, 1), ":", "Installing..")
		start := time.Now()
		if std, err := p.GoInstall(); err != nil {
			log.Println(pname(p.Name, 1), ":", fmt.Sprint(Red(err)), std)
			wr.Done()
		} else {
			log.Println(pname(p.Name, 5), ":", Green("Installed")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			if p.Run {
				runner := make(chan bool, 1)
				log.Println(pname(p.Name, 1), ":", "Running..")
				start = time.Now()
				go p.GoRun(channel, runner, wr)
				for {
					select {
					case <-runner:
						log.Println(pname(p.Name, 5), ":", Green("Has been run")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
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
		log.Println(pname(p.Name, 1), ":", "Building..")
		start := time.Now()
		if std, err := p.GoBuild(); err != nil {
			log.Println(pname(p.Name, 1), ":", fmt.Sprint(Red(err)), std)
		} else {
			log.Println(pname(p.Name, 5), ":", Green("Builded")+" after", MagentaS(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
		}
		return
	}
	return
}

// Fmt calls an implementation of the "gofmt"
func (p *Project) fmt(path string) error {
	if p.Fmt {
		if _, err := p.GoFmt(path); err != nil {
			log.Println(pname(p.Name, 1), Red("There are some GoFmt errors in "), ":", Magenta(path))
		}
	}
	return nil
}

// Cmd calls an wrapper for execute the commands after/before
func (p *Project) cmd(exit chan bool) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	cast := func(commands []string) {
		if errs := p.Cmd(commands); errs != nil {
			for _, err := range errs {
				log.Println(pname(p.Name, 2), Red(err))
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
		if _, err := p.GoTest(path); err != nil {
			log.Println(pname(p.Name, 1), Red("Go Test fails in "), ":", Magenta(path))
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
					fmt.Println(pname(p.Name, 1), ":", path)
				}
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
				if inArray(filepath.Ext(path), p.Watcher.Exts) {
					files++
					go func() {
						if err := p.fmt(path); err != nil {
							fmt.Println(err)
						}
					}()

				} else {
					folders++
					go func() {
						if err := p.test(path); err != nil {
							fmt.Println(err)
						}
					}()
				}
			}
		}
		return nil
	}

	if p.Path == "." || p.Path == "/" {
		p.base = wd
		p.Path = p.Wdir()
	} else if filepath.IsAbs(p.Path) {
		p.base = p.Path
	} else {
		p.base = filepath.Join(wd, p.Path)
	}

	for _, dir := range p.Watcher.Paths {
		base := filepath.Join(p.base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, walk); err != nil {
				log.Println(Red(err.Error()))
			}
		} else {
			return errors.New(base + " path doesn't exist")
		}
	}
	fmt.Println(pname(p.Name, 1), Red("Watching"), Magenta(files), "file/s", Magenta(folders), "folder/s")
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

// Routines launches the following methods: run, build, fmt, install
func (p *Project) routines(channel chan bool, wr *sync.WaitGroup) {
	wr.Add(1)
	go p.build()
	go p.install(channel, wr)
	wr.Wait()
}
