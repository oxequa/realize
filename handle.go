package core

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/oxequa/grace"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Watch paths and file extensions
type Watch struct {
	Deep bool     `yaml:"deep" json:"deep"`
	Path []string `yaml:"path,omitempty" json:"path,omitempty"`
	Ext  []string `yaml:"extension,omitempty" json:"extension,omitempty"`
}

// Logger events, output, errors
type Logger struct {
	Error   []interface{} `yaml:"error,omitempty" json:"error,omitempty"`
	Output  []interface{} `yaml:"output,omitempty" json:"output,omitempty"`
	General []interface{} `yaml:"general,omitempty" json:"general,omitempty"`
}

// Ignore paths and file extensions
type Ignore struct {
	Hidden bool     `yaml:"hidden,omitempty" json:"hidden,omitempty"`
	Ext    []string `yaml:"ext,omitempty" json:"ext,omitempty"`
	Path   []string `yaml:"path,omitempty" json:"path,omitempty"`
}

// Series list of commands to exec in sequence
type Series struct {
	Tasks []interface{} `yaml:"series,omitempty" json:"series,omitempty"`
}

// Parallel list of commands to exec in parallel
type Parallel struct {
	Tasks []interface{} `yaml:"parallel,omitempty" json:"parallel,omitempty"`
}

// Command instance, run from a custom path, log command output.
type Command struct {
	Log  bool   `yaml:"log,omitempty" json:"log,omitempty"`
	Task string `yaml:"task,omitempty" json:"task,omitempty"`
	Dir  string `yaml:"dir,omitempty" json:"dir,omitempty"`
}

// Response contains a command response
type Response struct {
	Cmd *Command
	Out string
	Err error
}

// Project struct contains all data about a program.
type Project struct {
	*Realize    `yaml:"-" json:"-"`
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	Logs        Logger            `yaml:"logs,omitempty" json:"logs,omitempty"`
	Watch       Watch             `yaml:"watch,omitempty" json:"watch,omitempty"`
	Ignore      Ignore            `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	Files       []string          `yaml:"-" json:"-"`
	Folders     []string          `yaml:"-" json:"-"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Tasks       []interface{}     `yaml:"tasks,omitempty" json:"tasks,omitempty"`
	TasksAfter  []interface{}     `yaml:"after,omitempty" json:"after,omitempty"`
	TasksBefore []interface{}     `yaml:"before,omitempty" json:"before,omitempty"`
}

// ToInterface convert an interface to an array of interface
func ToInterface(s interface{}) []interface{} {
	v := reflect.ValueOf(s)
	// There is no need to check, we want to panic if it's not slice or array
	intf := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		intf[i] = v.Index(i).Interface()
	}
	return intf
}

// Push a list of msg on stdout
func (p *Project) Push(msg ...interface{}) {
	if p.Realize != nil && len(p.Realize.Projects) > 1 {
		msg = append([]interface{}{Prefix(p.Name, White)}, msg...)
	}
	log.Println(msg...)
	if p.Realize != nil && p.Settings.Logs.File {
		f, err := os.OpenFile(logs, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		msg := append([]interface{}{time.Now().Format(time.RFC3339)}, msg...)
		if _, err = f.WriteString(fmt.Sprintln(msg...)); err != nil {
			panic(err)
		}
	}
}

// Recover check recover flag before push a msg
func (p *Project) Recover(msg ...interface{}) {
	if p.Realize != nil && p.Settings.Logs.Recovery {
		p.Push(msg...)
	}
}

// Scan an Project and wait for a change
func (p *Project) Scan(wg *sync.WaitGroup) (e error) {
	var ltime time.Time
	var w sync.WaitGroup
	var reload chan bool
	var watcher FileWatcher
	defer func() {
		close(reload)
		watcher.Close()
		grace.Recover(&e)
		wg.Done()
	}()
	// new chan
	reload = make(chan bool)
	// new file watcher
	watcher, err := NewFileWatcher(p.Settings.Polling)
	if err != nil {
		panic(e)
	}

	w.Add(1)
	// indexing
	go func() {
		defer w.Done()
		for _, item := range p.Watch.Path {
			abs, _ := filepath.Abs(item)
			glob, _ := filepath.Glob(abs)
			for _, g := range glob {
				if _, err := os.Stat(g); err == nil {
					if err = p.Walk(g, watcher); err != nil {
						p.Recover(Prefix("Indexed", Red), err.Error())
					}
				}
			}
		}
	}()
	// run tasks before
	p.Run(reload, p.TasksBefore)
	// wait indexing and before
	w.Wait()

	// run tasks list
	go p.Run(reload, p.Tasks...)
L:
	for {
		select {
		case event := <-watcher.Events():
			p.Recover(Prefix("Changed", Cyan), event.Name, event.Op.String())
			if time.Now().Truncate(time.Second).After(ltime) {
				switch event.Op {
				case fsnotify.Remove:
					watcher.Remove(event.Name)
					if s, _ := p.Validate(event.Name); s && Ext(event.Name) != "" {
						// stop and restart
						close(reload)
						reload = make(chan bool)
						p.Push(Prefix("Removed", Magenta), event.Name)
						go p.Run(reload, p.Tasks)
					}
				case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
					if s, fi := p.Validate(event.Name); s {
						if fi != nil && fi.IsDir() {
							if err = p.Walk(event.Name, watcher); err != nil {
								p.Recover(Prefix("Indexed", Red), err.Error())
							}
						} else {
							// stop and restart
							close(reload)
							reload = make(chan bool)
							p.Push(Prefix("Changed", Magenta), event.Name)
							go p.Run(reload, p.Tasks)
							ltime = time.Now().Truncate(time.Second)
						}
					}
				}
			}
		case err := <-watcher.Errors():
			p.Recover(Prefix("Watch", Red), err.Error())
		case <-p.Exit:
			// run task after
			p.Push(Prefix("Stopped", Red))
			p.Run(reload, p.TasksAfter...)
			break L
		}
	}
	return
}

// Walk file three
func (p *Project) Walk(path string, watcher FileWatcher) error {
	wdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	p.Push(Prefix("Watching", Green), Magenta.Regular(path))
	// Files walk
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !p.Watch.Deep {
			if len(p.Watch.Path) > 0 && !CheckInSlice(path, p.Watch.Path) {
				fi, _ := os.Stat(path)
				if fi.IsDir() {
					return filepath.SkipDir
				}
			}
		}
		if path == wdir || strings.HasPrefix(path, wdir) {
			if res, _ := p.Validate(path); res {
				fname := strings.Split(path, wdir)
				if fname[1] != "" {
					if Ext(fname[1]) == "" {
						p.Recover(Prefix("Indexed", Cyan), Magenta.Regular(fname[1]))
					} else {
						p.Recover(Prefix("Indexed", Cyan), fname[1])
					}
				}
				act := watcher.Walk(path, true)
				if ext := Ext(act); ext != "" {
					p.Files = append(p.Files, act)
				} else {
					p.Folders = append(p.Folders, act)
				}
			} else {
				fi, _ := os.Stat(path)
				if fi.IsDir() {
					return filepath.SkipDir
				}
			}
		}
		return nil
	})
	return nil
}

// Run exec a list of commands in parallel or in sequence
func (p *Project) Run(reload <-chan bool, tasks ...interface{}) {
	var w sync.WaitGroup
	// Loop tasks
	for _, task := range tasks {
		switch t := task.(type) {
		case Command:
			select {
			case <-reload:
				p.Recover(Prefix("Tasks loop stopped", Red))
				w.Done()
				break
			default:
				// Exec command
				if len(t.Task) > 0 {
					p.Recover(Prefix("Running task", Green), t.Task)
					w.Add(1)
					p.Exec(t, &w, reload)
				}
			}
			break
		case Parallel:
			var wl sync.WaitGroup
			for _, t := range t.Tasks {
				wl.Add(1)
				go func(t interface{}) {
					p.Run(reload, t)
					wl.Done()
				}(t)
			}
			wl.Wait()
			break
		case Series:
			for _, c := range t.Tasks {
				p.Run(reload, c)
			}
			break
		}
	}
	w.Wait()
}

// Validate a path
func (p *Project) Validate(path string) (s bool, fi os.FileInfo) {
	// check temp file
	if TempFile(path) {
		return
	}
	// validate hidden
	if p.Ignore.Hidden {
		if Hidden(path) {
			return
		}
	}
	// validate extension
	if e := Ext(path); e != "" {
		if len(p.Ignore.Ext) > 0 {
			for _, v := range p.Ignore.Ext {
				if v == e {
					return
				}
			}
		}
		if len(p.Watch.Ext) > 0 {
			match := false
			for _, v := range p.Watch.Ext {
				if v == e {
					match = true
					break
				}
			}
			if !match {
				return
			}
		}
	}
	// validate path
	if fpath, err := filepath.Abs(path); err != nil {
		p.Recover(Prefix("Error", Red), err.Error())
		return
	} else {
		if len(p.Ignore.Path) > 0 {
			for _, v := range p.Ignore.Path {
				v, _ := filepath.Abs(v)
				if strings.Contains(fpath, v) {
					return
				}
				if strings.Contains(v, "*") {
					// check glob
					paths, err := filepath.Glob(v)
					if err != nil {
						p.Recover(Prefix("Error", Red), err.Error())
						return
					}
					for _, p := range paths {
						if strings.Contains(p, fpath) {
							return
						}
					}
				}
			}
		}
		if len(p.Watch.Path) > 0 {
			match := false
			for _, v := range p.Watch.Path {
				v, _ := filepath.Abs(v)
				if strings.Contains(fpath, v) {
					match = true
					break
				}
				if strings.Contains(v, "*") {
					// check glob
					paths, err := filepath.Glob(v)
					if err != nil {
						p.Recover(Prefix("Error", Red), err.Error())
						return
					}
					for _, p := range paths {
						if strings.Contains(p, fpath) {
							match = true
							break
						}
					}
				}
			}
			if !match {
				return
			}
		}
	}
	s = true
	return
}

// Exec a command
func (p *Project) Exec(c Command, w *sync.WaitGroup, reload <-chan bool) error {
	var build *exec.Cmd
	var lifetime time.Time
	defer func() {
		// ref https://github.com/golang/go/issues/5615
		// ref https://github.com/golang/go/issues/6720
		if build != nil {
			if runtime.GOOS == "windows" {
				build.Process.Kill()
				build.Process.Wait()
			} else {
				build.Process.Signal(os.Interrupt)
				build.Process.Wait()
			}
		}
		// Print command end
		p.Push(Prefix("Cmd", Green),
			Print("End",
				Green.Regular("'")+strings.Split(c.Task, " -")[0]+Green.Regular("'"), "in",
				Magenta.Regular(big.NewFloat(time.Since(lifetime).Seconds()).Text('f', 3), "s")))
		// Command done
		w.Done()
	}()
	done := make(chan error)
	// Split command
	args := strings.Split(c.Task, " ")
	build = exec.Command(args[0], args[1:]...)
	//TODO Custom error pattern

	// Get exec dir
	if len(c.Dir) > 0 {
		build.Dir = c.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		build.Dir = dir
	}
	// stdout
	stdout, err := build.StdoutPipe()
	if err != nil {
		return err
	}
	// stderr
	stderr, err := build.StderrPipe()
	if err != nil {
		return err
	}
	// Start command
	if err := build.Start(); err != nil {
		return err
	} else {
		// Print command start
		p.Push(Prefix("Cmd", Green),
			Print("Start",
				Green.Regular("'")+strings.Split(c.Task, " -")[0]+Green.Regular("'")))
		// Start time
		lifetime = time.Now()
	}
	// Scan outputs and errors generated by command exec
	exOut, exErr := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOut, stopErr := make(chan bool, 1), make(chan bool, 1)
	scanner := func(output *bufio.Scanner, end chan bool, err bool) {
		for output.Scan() {
			if len(output.Text()) > 0 {
				if err {
					//TODO check custom error pattern
					p.Push(Prefix("Err", Red), output.Text())
				} else {
					p.Push(Prefix("Out", Blue), output.Text())
				}
			}
		}
		close(end)
	}
	// Wait command end
	go func() { done <- build.Wait() }()
	// Run scanner
	go scanner(exErr, stopErr, true)
	go scanner(exOut, stopOut, false)

	// Wait command result
	select {
	case <-reload:
		p.Recover(Prefix("Build process stopped", Yellow))
		// Stop running command
		build.Process.Kill()
		break
	case <-done:
		p.Recover(Prefix("Build done", Green))
		break
	}
	return nil
}
