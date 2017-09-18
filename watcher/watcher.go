package watcher

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/tockins/realize/settings"
	"github.com/tockins/realize/style"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var msg string
var out BufferOut

// Project defines the informations of a single project
type Project struct {
	settings.Settings  `yaml:"-" json:"-"`
	parent             *Blueprint
	path               string
	tools              tools
	base               string
	paths              []string
	lastChangedOn      time.Time
	Name               string            `yaml:"name" json:"name"`
	Path               string            `yaml:"path" json:"path"`
	Environment        map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Cmds               Cmds              `yaml:"commands" json:"commands"`
	Args               []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Watcher            Watcher           `yaml:"watcher" json:"watcher"`
	Buffer             Buffer            `yaml:"-" json:"buffer"`
	ErrorOutputPattern string            `yaml:"errorOutputPattern,omitempty" json:"errorOutputPattern,omitempty"`
}

// Watch the project by fsnotify
func (p *Project) watchByNotify() {
	wr := sync.WaitGroup{}
	channel := make(chan bool, 1)
	watcher := &fsnotify.Watcher{}
	exit := make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	watcher, err := fsnotify.NewWatcher()
	p.Fatal(err)
	defer func() {
		p.cmd("after", true)
		wg.Done()
	}()
	p.cmd("before", true)
	go p.routines(&wr, channel, watcher, "")
	p.lastChangedOn = time.Now().Truncate(time.Second)
L:
	for {
		select {
		case event := <-watcher.Events:
			if time.Now().Truncate(time.Second).After(p.lastChangedOn) {
				p.lastChangedOn = time.Now().Truncate(time.Second)
				if file, err := os.Lstat(event.Name); err == nil {
					if file.Size() > 0 {
						p.lastChangedOn = time.Now().Truncate(time.Second)
						ext := filepath.Ext(event.Name)
						if inArray(ext, p.Watcher.Exts) {
							if p.Cmds.Run {
								close(channel)
								channel = make(chan bool)
							}
							// repeat the initial cycle
							msg = fmt.Sprintln(p.pname(p.Name, 4), ":", style.Magenta.Bold(strings.ToUpper(ext[1:])+" changed"), style.Magenta.Bold(event.Name))
							out = BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + event.Name}
							p.stamp("log", out, msg, "")
							// check if is deleted
							if event.Op&fsnotify.Remove == fsnotify.Remove {
								watcher.Remove(event.Name)
								go p.routines(&wr, channel, watcher, "")
							} else {
								go p.routines(&wr, channel, watcher, event.Name)
							}
							p.lastChangedOn = time.Now().Truncate(time.Second)
						}
					}
				}
			}
		case err := <-watcher.Errors:
			msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Regular(err.Error()))
			out = BufferOut{Time: time.Now(), Text: err.Error()}
			p.stamp("error", out, msg, "")
		case <-exit:
			break L
		}
	}
	return
}

// Watch the project by polling
func (p *Project) watchByPolling() {
	wr := sync.WaitGroup{}
	watcher := new(pollWatcher)
	channel := make(chan bool, 1)
	exit := make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	defer func() {
		p.cmd("after", true)
		wg.Done()
	}()
	p.cmd("before", true)
	go p.routines(&wr, channel, watcher, "")
	p.lastChangedOn = time.Now().Truncate(time.Second)
	walk := func(changed string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if !watcher.isWatching(changed) {
			return nil
		} else if !info.ModTime().Truncate(time.Second).After(p.lastChangedOn) {
			return nil
		}
		if index := strings.Index(filepath.Ext(changed), "__"); index != -1 {
			return nil
		}
		ext := filepath.Ext(changed)
		if inArray(ext, p.Watcher.Exts) {
			if p.Cmds.Run {
				close(channel)
				channel = make(chan bool)
			}
			p.lastChangedOn = time.Now().Truncate(time.Second)
			// repeat the initial cycle
			msg = fmt.Sprintln(p.pname(p.Name, 4), ":", style.Magenta.Bold(strings.ToUpper(ext[1:])+" changed"), style.Magenta.Bold(changed))
			out = BufferOut{Time: time.Now(), Text: strings.ToUpper(ext[1:]) + " changed " + changed}
			p.stamp("log", out, msg, "")
			go p.routines(&wr, channel, watcher, changed)
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
					p.stamp("error", out, msg, "")
				}
			} else {
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", base, "path doesn't exist")
				out = BufferOut{Time: time.Now(), Text: base + " path doesn't exist"}
				p.stamp("error", out, msg, "")
			}
			select {
			case <-exit:
				return
			case <-time.After(p.parent.Legacy.Interval / time.Duration(len(p.Watcher.Paths))):
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
			p.stamp("error", out, msg, stream)
		} else {
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Built")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			out = BufferOut{Time: time.Now(), Text: "Built after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
			p.stamp("log", out, msg, stream)
		}
		return err
	}
	return nil
}

// Install calls an implementation of "go install"
func (p *Project) install() error {
	if p.Cmds.Install.Status {
		start := time.Now()
		log.Println(p.pname(p.Name, 1), ":", "Installing..")
		stream, err := p.goInstall()
		if err != nil {
			msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Bold("Go Install"), style.Red.Regular(err.Error()))
			out = BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Install", Stream: stream}
			p.stamp("error", out, msg, stream)
		} else {
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Installed")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
			out = BufferOut{Time: time.Now(), Text: "Installed after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
			p.stamp("log", out, msg, stream)
		}
		return err
	}
	return nil
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

// Watch the files tree of a project
func (p *Project) walk(watcher watcher) error {
	var files, folders int64
	walk := func(path string, info os.FileInfo, err error) error {
		if !p.ignore(path) {
			if ((info.IsDir() && len(filepath.Ext(path)) == 0 && !strings.HasPrefix(path, ".")) && !strings.Contains(path, "/.")) || (inArray(filepath.Ext(path), p.Watcher.Exts)) {
				if p.Watcher.Preview {
					log.Println(p.pname(p.Name, 1), ":", path)
				}
				if err = watcher.Add(path); err != nil {
					return filepath.SkipDir
				}
				if inArray(filepath.Ext(path), p.Watcher.Exts) {
					p.paths = append(p.paths, path)
					files++
				} else {
					folders++
				}
			}
		}
		return nil
	}

	for _, dir := range p.Watcher.Paths {
		base := filepath.Join(p.base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, walk); err != nil {
				log.Println(style.Red.Bold(err.Error()))
				p.tool(base, p.tools.Fmt)
				p.tool(base, p.tools.Vet)
				p.tool(base, p.tools.Test)
				p.tool(base, p.tools.Generate)
			}
		} else {
			return errors.New(base + " path doesn't exist")
		}
	}
	msg = fmt.Sprintln(p.pname(p.Name, 1), ":", style.Blue.Bold("Watching"), style.Magenta.Bold(files), "file/s", style.Magenta.Bold(folders), "folder/s")
	out = BufferOut{Time: time.Now(), Text: "Watching " + strconv.FormatInt(files, 10) + " files/s " + strconv.FormatInt(folders, 10) + " folder/s"}
	p.stamp("log", out, msg, "")
	return nil
}

// Cmd calls an wrapper for execute the commands after/before
func (p *Project) cmd(flag string, global bool) {
	for _, cmd := range p.Watcher.Scripts {
		if strings.ToLower(cmd.Type) == flag && cmd.Global == global {
			err, logs := p.command(cmd)
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Bold("Command"), style.Green.Bold("\"")+cmd.Command+style.Green.Bold("\""))
			out = BufferOut{Time: time.Now(), Text: cmd.Command, Type: flag}
			if err != "" {
				p.stamp("error", out, msg, "")
			} else {
				p.stamp("log", out, msg, "")
			}
			if logs != "" && cmd.Output {
				msg = fmt.Sprintln(logs)
				out = BufferOut{Time: time.Now(), Text: logs, Type: flag}
				p.stamp("log", out, "", msg)
			}
			if err != "" {
				msg = fmt.Sprintln(style.Red.Regular(err))
				out = BufferOut{Time: time.Now(), Text: err, Type: flag}
				p.stamp("error", out, "", msg)
			}
		}
	}
}

//  Tool logs the result of a go command
func (p *Project) tool(path string, tool tool) error {
	if tool.status {
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "") {
			if strings.HasSuffix(path, ".go") {
				tool.options = append(tool.options, path)
				path = p.base
			}
			if stream, err := p.goTool(path, tool.cmd, tool.options...); err != nil {
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", style.Red.Bold(tool.name), style.Red.Regular("there are some errors in"), ":", style.Magenta.Bold(path))
				out = BufferOut{Time: time.Now(), Text: "there are some errors in", Path: path, Type: tool.name, Stream: stream}
				p.stamp("error", out, msg, stream)
				return err
			}
		}
	}
	return nil
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
				msg = fmt.Sprintln(p.pname(p.Name, 5), ":", style.Green.Regular("Started")+" after", style.Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
				out = BufferOut{Time: time.Now(), Text: "Started after " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
				p.stamp("log", out, msg, "")
				return
			}
		}
	}
}

// Print on files, cli, ws
func (p *Project) stamp(t string, o BufferOut, msg string, stream string) {
	switch t {
	case "out":
		p.Buffer.StdOut = append(p.Buffer.StdOut, o)
		if p.Files.Outputs.Status {
			f := p.Create(p.base, p.Files.Outputs.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
	case "log":
		p.Buffer.StdLog = append(p.Buffer.StdLog, o)
		if p.Files.Logs.Status {
			f := p.Create(p.base, p.Files.Logs.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
	case "error":
		p.Buffer.StdErr = append(p.Buffer.StdErr, o)
		if p.Files.Errors.Status {
			f := p.Create(p.base, p.Files.Errors.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.Fatal(err, "")
			}
		}
	}
	if msg != "" {
		log.Print(msg)
	}
	if stream != "" {
		fmt.Fprint(style.Output, stream)
	}
	go func() {
		p.parent.Sync <- "sync"
	}()
}

// Routines launches the toolchain run, build, install
func (p *Project) routines(wr *sync.WaitGroup, channel chan bool, watcher watcher, file string) {
	p.cmd("before", false)
	if len(file) > 0 {
		path := filepath.Dir(file)
		p.tool(file, p.tools.Fmt)
		p.tool(path, p.tools.Vet)
		p.tool(path, p.tools.Test)
		p.tool(path, p.tools.Generate)
	} else {
		p.Fatal(p.walk(watcher))
	}
	install := p.install()
	build := p.build()
	wr.Add(1)
	if install == nil && build == nil {
		go p.run(channel, wr)
	}
	wr.Wait()
	if len(file) > 0 {
		p.cmd("after", false)
	}
}

// Add a path to paths list
func (w *pollWatcher) Add(path string) error {
	if w.paths == nil {
		w.paths = map[string]bool{}
	}
	w.paths[path] = true
	return nil
}

// Check if is watching
func (w *pollWatcher) isWatching(path string) bool {
	a, b := w.paths[path]
	return a && b
}
