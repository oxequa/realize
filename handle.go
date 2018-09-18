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
	Ext  []string `yaml:"ext,omitempty" json:"ext,omitempty"`
	Path []string `yaml:"path,omitempty" json:"path,omitempty"`
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
	Tasks []interface{} `yaml:"sequence,omitempty" json:"sequence,omitempty"`
}

// Command fields. Path run from a custom path. Log display command output.
type Command struct {
	Log bool   `yaml:"log,omitempty" json:"log,omitempty"`
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Dir string `yaml:"dir,omitempty" json:"dir,omitempty"`
}

// Response contains a command response
type Response struct {
	Cmd *Command
	Out string
	Err error
}

// Activity struct contains all data about a program.
type Activity struct {
	*Realize
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	Logs        Logger            `yaml:"logs,omitempty" json:"logs,omitempty"`
	Watch       *Watch            `yaml:"watch,omitempty" json:"watch,omitempty"`
	Ignore      *Ignore           `yaml:"ignore,omitempty" json:"ignore,omitempty"`
	Files       []string          `yaml:"files,omitempty" json:"files,omitempty"`
	Folders     []string          `yaml:"folders,omitempty" json:"folders,omitempty"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Tasks       []interface{}     `yaml:"tasks,omitempty" json:"tasks,omitempty"`
	TasksAfter  []interface{}     `yaml:"after,omitempty" json:"after,omitempty"`
	TasksBefore []interface{}     `yaml:"before,omitempty" json:"before,omitempty"`
}

// Parallel list of commands to exec in parallel
type Parallel struct {
	Tasks []interface{} `yaml:"parallel,omitempty" json:"parallel,omitempty"`
}

// ToInterface convert an interface to an array of interface
func toInterface(s interface{}) []interface{} {
	v := reflect.ValueOf(s)
	// There is no need to check, we want to panic if it's not slice or array
	intf := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		intf[i] = v.Index(i).Interface()
	}
	return intf
}

// Push a list of msg on stdout
func (a *Activity) Push(msg ...interface{}) {
	if a.Realize != nil && len(a.Realize.Schema) > 1 {
		msg = append([]interface{}{Prefix(a.Name, White)}, msg...)
	}
	log.Println(msg...)
	if a.Realize != nil && a.Settings.Broker.File {
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
func (a *Activity) Recover(msg ...interface{}) {
	if a.Realize != nil && a.Settings.Broker.Recovery {
		a.Push(msg...)
	}
}

// Scan an activity and wait for a change
func (a *Activity) Scan(wg *sync.WaitGroup) (e error) {
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
	watcher, err := NewFileWatcher(a.Settings.Polling)
	if err != nil {
		panic(e)
	}

	w.Add(1)
	// indexing
	go func() {
		defer w.Done()
		for _, p := range a.Watch.Path {
			abs, _ := filepath.Abs(p)
			glob, _ := filepath.Glob(abs)
			for _, g := range glob {
				if _, err := os.Stat(g); err == nil {
					if err = a.Walk(g, watcher); err != nil {
						a.Recover(Prefix("Indexing", Red), err.Error())
					}
				}
			}
		}
	}()
	// run tasks before
	a.Run(reload, a.TasksBefore)
	// wait indexing and before
	w.Wait()

	// run tasks list
	go a.Run(reload, a.Tasks...)
L:
	for {
		select {
		case event := <-watcher.Events():
			a.Recover(Prefix("File Changed", Magenta), event.Name)
			if time.Now().Truncate(time.Second).After(ltime) {
				switch event.Op {
				case fsnotify.Remove:
					watcher.Remove(event.Name)
					if s, _ := a.Validate(event.Name); s && Ext(event.Name) != "" {
						// stop and restart
						close(reload)
						reload = make(chan bool)
						a.Push(Prefix("Removed", Magenta), event.Name)
						go a.Run(reload, a.Tasks)
					}
				case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
					if s, fi := a.Validate(event.Name); s {
						if fi.IsDir() {
							if err = a.Walk(event.Name, watcher); err != nil {
								a.Recover(Prefix("Indexing", Red), err.Error())
							}
						} else {
							// stop and restart
							close(reload)
							reload = make(chan bool)
							a.Push(Prefix("Changed", Magenta), event.Name)
							go a.Run(reload, a.Tasks)
							ltime = time.Now().Truncate(time.Second)
						}
					}
				}
			}
		case err := <-watcher.Errors():
			a.Recover(Prefix("Watch", Red), err.Error())
		case <-a.Exit:
			// run task after
			a.Recover(Prefix("Loop stopped", Red))
			a.Run(reload, a.TasksAfter)
			break L
		}
	}
	return
}

// Walk file three
func (a *Activity) Walk(path string, watcher FileWatcher) error {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		wdir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		if path == wdir || strings.HasPrefix(path, wdir) {
			if res, _ := a.Validate(path); res {
				a.Recover(Prefix("Indexing", Magenta), path)
				act := watcher.Walk(path, true)
				if ext := Ext(act); ext != "" {
					a.Files = append(a.Files, act)
				} else {
					a.Folders = append(a.Files, act)
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
func (a *Activity) Run(reload <-chan bool, tasks ...interface{}) {
	var w sync.WaitGroup
	// Loop tasks
	for _, task := range tasks {
		switch t := task.(type) {
		case Command:
			select {
			case <-reload:
				a.Recover(Prefix("Tasks loop stopped", Red))
				w.Done()
				break
			default:
				// Exec command
				if len(t.Cmd) > 0 {
					a.Recover(Prefix("Running task", Green), t.Cmd)
					w.Add(1)
					a.Exec(t, &w, reload)
				}
			}
			break
		case Parallel:
			var wl sync.WaitGroup
			for _, t := range t.Tasks {
				wl.Add(1)
				go func(t interface{}) {
					a.Run(reload, t)
					wl.Done()
				}(t)
			}
			wl.Wait()
			break
		case Series:
			for _, c := range t.Tasks {
				a.Run(reload, c)
			}
			break
		}
	}
	w.Wait()
}

// Validate a path
func (a *Activity) Validate(path string) (s bool, fi os.FileInfo) {
	if len(path) == 0 {
		return
	}
	// validate hidden
	if a.Ignore != nil && a.Ignore.Hidden {
		if Hidden(path) {
			return
		}
	}
	// validate extension
	if e := Ext(path); e != "" {
		if a.Ignore != nil && len(a.Ignore.Ext) > 0 {
			for _, v := range a.Ignore.Ext {
				if v == e {
					return
				}
			}
		}
		if a.Watch != nil && len(a.Watch.Ext) > 0 {
			match := false
			for _, v := range a.Watch.Ext {
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
		a.Recover(Prefix("Error", Red), err.Error())
		return
	} else {
		if a.Ignore != nil && len(a.Ignore.Path) > 0 {
			for _, v := range a.Ignore.Path {
				v, _ := filepath.Abs(v)
				if strings.Contains(fpath, v) {
					return
				}
				if strings.Contains(v, "*") {
					// check glob
					paths, err := filepath.Glob(v)
					if err != nil {
						a.Recover(Prefix("Error", Red), err.Error())
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
		if a.Watch != nil && len(a.Watch.Path) > 0 {
			match := false
			for _, v := range a.Watch.Path {
				v, _ := filepath.Abs(v)
				if strings.Contains(fpath, v) {
					match = true
					break
				}
				if strings.Contains(v, "*") {
					// check glob
					paths, err := filepath.Glob(v)
					if err != nil {
						a.Recover(Prefix("Error", Red), err.Error())
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
func (a *Activity) Exec(c Command, w *sync.WaitGroup, reload <-chan bool) error {
	var build *exec.Cmd
	var lifetime time.Time
	defer func() {
		// https://github.com/golang/go/issues/5615
		// https://github.com/golang/go/issues/6720
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
		a.Push(Prefix("Cmd", Green),
			Print("Finished",
				Green.Regular("'")+
					strings.Split(c.Cmd, " -")[0]+
					Green.Regular("'"),
				"in", Magenta.Regular(big.NewFloat(time.Since(lifetime).Seconds()).Text('f', 3), "s")))
		// Command done
		w.Done()
	}()
	done := make(chan error)
	// Split command
	args := strings.Split(c.Cmd, " ")
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
		a.Push(Prefix("Cmd", Green),
			Print("Running\t",
				Green.Regular("'")+
					strings.Split(c.Cmd, " -")[0]+
					Green.Regular("'")))
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
					a.Push(Prefix("Err", Red), output.Text())
				} else {
					a.Push(Prefix("Out", Blue), output.Text())
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
		a.Recover(Prefix("Build process stopped", Yellow))
		// Stop running command
		build.Process.Kill()
		break
	case <-done:
		a.Recover(Prefix("Build done", Green))
		break
	}
	return nil
}
