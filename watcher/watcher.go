package watcher

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/tockins/realize/style"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var msg string
var out BufferOut

func (w *pollWatcher) Add(path string) error {
	if w.paths == nil {
		w.paths = map[string]bool{}
	}
	w.paths[path] = true
	return nil
}

func (w *pollWatcher) isWatching(path string) bool {
	a, b := w.paths[path]
	return a && b
}

// Watch the project by polling
func (p *Project) watchByPolling() {
	var wr sync.WaitGroup
	var watcher = new(pollWatcher)
	channel, exit := make(chan bool, 1), make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	defer func() {
		p.cmd("after")
		wg.Done()
	}()

	p.cmd("before")
	p.Fatal(p.watch(watcher))
	go p.routines(channel, &wr)
	p.LastChangedOn = time.Now().Truncate(time.Second)
	walk := func(changed string, info os.FileInfo, err error) error {
		var ext string
		if err != nil {
			return err
		} else if !watcher.isWatching(changed) {
			return nil
		} else if !info.ModTime().Truncate(time.Second).After(p.LastChangedOn) {
			return nil
		}
		if index := strings.Index(filepath.Ext(changed), "_"); index == -1 {
			ext = filepath.Ext(changed)
		} else {
			ext = filepath.Ext(changed)[0:index]
		}
		i := strings.Index(changed, filepath.Ext(changed))
		file := changed[:i] + ext
		path := filepath.Dir(changed[:i])
		if changed[:i] != "" && inArray(ext, p.Watcher.Exts) {
			if p.Cmds.Run {
				close(channel)
				channel = make(chan bool)
			}
			p.LastChangedOn = time.Now().Truncate(time.Second)
			// repeat the initial cycle
			msg = fmt.Sprintln(p.pname(p.Name, 4), ":", style.Magenta.Bold(strings.ToUpper(ext[1:])+" changed"), style.Magenta.Bold(file))
			out = BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + file}
			p.print("log", out, msg, "")

			p.cmd("change")
			p.tool(file, p.tools.Fmt)
			p.tool(file, p.tools.Vet)
			p.tool(path, p.tools.Test)
			p.tool(path, p.tools.Generate)
			go p.routines(channel, &wr)
		}
		return nil
	}
	for {
		for _, dir := range p.Watcher.Paths {
			base := filepath.Join(p.base, dir)
			if _, err := os.Stat(base); err == nil {
				if err := filepath.Walk(base, walk); err != nil {
					msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Regular(err.Error()))
					out = BufferOut{Time: time.Now(), Text: err.Error()}
					p.print("error", out, msg, "")
				}
			} else {
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", base, "path doesn't exist")
				out = BufferOut{Time: time.Now(), Text: base + " path doesn't exist"}
				p.print("error", out, msg, "")
			}
			select {
			case <-exit:
				return
			case <-time.After(p.parent.Legacy.Interval / time.Duration(len(p.Watcher.Paths))):
			}
		}
	}
}

// Watch the project by fsnotify
func (p *Project) watchByNotify() {
	var wr sync.WaitGroup
	var watcher *fsnotify.Watcher
	channel, exit := make(chan bool, 1), make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	watcher, err := fsnotify.NewWatcher()
	p.Fatal(err)
	defer func() {
		p.cmd("after")
		wg.Done()
	}()

	p.cmd("before")
	p.Fatal(p.watch(watcher))
	go p.routines(channel, &wr)
	p.LastChangedOn = time.Now().Truncate(time.Second)
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
						if p.Cmds.Run {
							close(channel)
							channel = make(chan bool)
						}
						// repeat the initial cycle
						msg = fmt.Sprintln(p.pname(p.Name, 4), ":", style.Magenta.Bold(strings.ToUpper(ext[1:])+" changed"), style.Magenta.Bold(file))
						out = BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + file}
						p.print("log", out, msg, "")

						p.cmd("change")
						p.tool(file, p.tools.Fmt)
						p.tool(path, p.tools.Vet)
						p.tool(path, p.tools.Test)
						p.tool(path, p.tools.Generate)
						go p.routines(channel, &wr)
						p.LastChangedOn = time.Now().Truncate(time.Second)
					}
				}
			}
		case err := <-watcher.Errors:
			msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Regular(err.Error()))
			out = BufferOut{Time: time.Now(), Text: err.Error()}
			p.print("error", out, msg, "")
		case <-exit:
			return
		}
	}
}

// Watch the files tree of a project
func (p *Project) watch(watcher watcher) error {
	var files, folders int64
	wd, _ := os.Getwd()
	walk := func(path string, info os.FileInfo, err error) error {
		if !p.ignore(path) {
			if (info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.HasPrefix(path, ".")) && !strings.Contains(path, "/.") || (inArray(filepath.Ext(path), p.Watcher.Exts)) {
				if p.Watcher.Preview {
					log.Println(p.pname(p.Name, 1), ":", path)
				}
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
				if inArray(filepath.Ext(path), p.Watcher.Exts) {
					files++
					p.tool(path, p.tools.Fmt)
				} else {
					folders++
					p.tool(path, p.tools.Vet)
					p.tool(path, p.tools.Test)
					p.tool(path, p.tools.Generate)
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
				log.Println(style.Red.Bold(err.Error()))
			}
		} else {
			return errors.New(base + " path doesn't exist")
		}
	}
	msg = fmt.Sprintln(p.pname(p.Name, 1), ":", style.Blue.Bold("Watching"), style.Magenta.Bold(files), "file/s", style.Magenta.Bold(folders), "folder/s")
	out = BufferOut{Time: time.Now(), Text: "Watching " + strconv.FormatInt(files, 10) + " files/s " + strconv.FormatInt(folders, 10) + " folder/s"}
	p.print("log", out, msg, "")
	return nil
}

// Install calls an implementation of "go install"
func (p *Project) install() error {
	if p.Cmds.Bin.Status {
		start := time.Now()
		log.Println(p.pname(p.Name, 1), ":", "Installing..")
		stream, err := p.goInstall()
		if err != nil {
			msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Bold("Go Install"), style.Red.Regular(err.Error()))
			out = BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Install", Stream: stream}
			p.print("error", out, msg, stream)
		} else {
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Installed")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			out = BufferOut{Time: time.Now(), Text: "Installed after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
			p.print("log", out, msg, stream)
		}
		return err
	}
	return nil
}

// Install calls an implementation of "go run"
func (p *Project) run(channel chan bool, wr *sync.WaitGroup) {
	if p.Cmds.Run {
		start := time.Now()
		runner := make(chan bool, 1)
		log.Println(p.pname(p.Name, 1), ":", "Running..")
		go p.goRun(channel, runner, wr)
		for {
			select {
			case <-runner:
				msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Has been run")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
				out = BufferOut{Time: time.Now(), Text: "Has been run after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
				p.print("log", out, msg, "")
				return
			}
		}
	}
}

// Build calls an implementation of the "go build"
func (p *Project) build() error {
	if p.Cmds.Build.Status {
		start := time.Now()
		log.Println(p.pname(p.Name, 1), ":", "Building..")
		stream, err := p.goBuild()
		if err != nil {
			msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Bold("Go Build"), style.Red.Regular(err.Error()))
			out = BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Build", Stream: stream}
			p.print("error", out, msg, stream)
		} else {
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Builded")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			out = BufferOut{Time: time.Now(), Text: "Builded after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
			p.print("log", out, msg, stream)
		}
		return err
	}
	return nil
}

func (p *Project) tool(path string, tool tool) error {
	if tool.status != nil {
		v := reflect.ValueOf(tool.status).Elem()
		if v.Interface().(bool) && (strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "")) {
			if strings.HasSuffix(path, ".go") {
				tool.options = append(tool.options, path)
				path = p.base
			}
			if stream, err := p.goTools(path, tool.cmd, tool.options...); err != nil {
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Bold(tool.name), style.Red.Regular("there are some errors in"), ":", style.Magenta.Bold(path))
				out = BufferOut{Time: time.Now(), Text: "there are some errors in", Path: path, Type: tool.name, Stream: stream}
				p.print("error", out, msg, stream)
				return err
			}
		}
	}
	return nil
}

// Cmd calls an wrapper for execute the commands after/before
func (p *Project) cmd(flag string) {
	for _, cmd := range p.Watcher.Scripts {
		if strings.ToLower(cmd.Type) == flag {
			errors, logs := p.command(cmd)
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Bold("Command"), style.Green.Bold("\"")+cmd.Command+style.Green.Bold("\""))
			out = BufferOut{Time: time.Now(), Text: cmd.Command, Type: flag}
			if logs != "" {
				p.print("log", out, msg, "")
			}
			if errors != "" {
				p.print("error", out, msg, "")
			}
			if logs != "" {
				msg = fmt.Sprintln(logs)
				out = BufferOut{Time: time.Now(), Text: logs, Type: flag}
				p.print("log", out, "", msg)
			}
			if errors != "" {
				msg = fmt.Sprintln(style.Red.Regular(errors))
				out = BufferOut{Time: time.Now(), Text: errors, Type: flag}
				p.print("error", out, "", msg)
			}
		}
	}
}

// Ignore and validate a path
func (p *Project) ignore(str string) bool {
	for _, v := range p.Watcher.Ignore {
		if strings.Contains(str, filepath.Join(p.base, v)) {
			return true
		}
	}
	return false
}

// Routines launches the toolchain run, build, install
func (p *Project) routines(channel chan bool, wr *sync.WaitGroup) {
	install := p.install()
	build := p.build()
	wr.Add(1)
	if install == nil && build == nil {
		go p.run(channel, wr)
	}
	wr.Wait()
}

// Defines the colors scheme for the project name
func (p *Project) pname(name string, color int) string {
	switch color {
	case 1:
		name = style.Yellow.Regular("[") + strings.ToUpper(name) + style.Yellow.Regular("]")
		break
	case 2:
		name = style.Yellow.Regular("[") + style.Red.Bold(strings.ToUpper(name)) + style.Yellow.Regular("]")
		break
	case 3:
		name = style.Yellow.Regular("[") + style.Blue.Bold(strings.ToUpper(name)) + style.Yellow.Regular("]")
		break
	case 4:
		name = style.Yellow.Regular("[") + style.Magenta.Bold(strings.ToUpper(name)) + style.Yellow.Regular("]")
		break
	case 5:
		name = style.Yellow.Regular("[") + style.Green.Bold(strings.ToUpper(name)) + style.Yellow.Regular("]")
		break
	}
	return name
}

// Print on files, cli, ws
func (p *Project) print(t string, o BufferOut, msg string, stream string) {
	switch t {
	case "out":
		p.Buffer.StdOut = append(p.Buffer.StdOut, o)
		if p.Streams.FileOut {
			f := p.Create(p.base, p.parent.Resources.Outputs)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
	case "log":
		p.Buffer.StdLog = append(p.Buffer.StdLog, o)
		if p.Streams.FileLog {
			f := p.Create(p.base, p.parent.Resources.Logs)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
		if msg != "" {
			log.Print(msg)
		}
	case "error":
		p.Buffer.StdErr = append(p.Buffer.StdErr, o)
		if p.Streams.FileErr {
			f := p.Create(p.base, p.parent.Resources.Errors)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
		if msg != "" {
			log.Print(msg)
		}
	}
	if stream != "" {
		fmt.Print(stream)
	}
	go func() {
		p.parent.Sync <- "sync"
	}()
}
