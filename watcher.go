package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
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

var (
	msg string
	out BufferOut
	wg  sync.WaitGroup
)

const (
	msgStop    = "killed"
	extWindows = extWindows
)

// Watch struct defines options for livereload
type Watch struct {
	Preview bool      `yaml:"preview,omitempty" json:"preview,omitempty"`
	Paths   []string  `yaml:"paths" json:"paths"`
	Exts    []string  `yaml:"extensions" json:"extensions"`
	Ignore  []string  `yaml:"ignored_paths,omitempty" json:"ignored_paths,omitempty"`
	Scripts []Command `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// Result channel with data stream and errors
type Result struct {
	stream string
	err    error
}

// Buffer define an array buffer for each log files
type Buffer struct {
	StdOut []BufferOut `json:"stdOut"`
	StdLog []BufferOut `json:"stdLog"`
	StdErr []BufferOut `json:"stdErr"`
}

// Project defines the informations of a single project
type Project struct {
	parent               *realize
	watcher              FileWatcher
	init                 bool
	files, folders       int64
	base, path, lastFile string
	tools                []tool
	paths                []string
	lastTime             time.Time
	Settings             `yaml:"-" json:"-"`
	Name                 string            `yaml:"name" json:"name"`
	Path                 string            `yaml:"path" json:"path"`
	Environment          map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Cmds                 Cmds              `yaml:"commands" json:"commands"`
	Args                 []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Watcher              Watch             `yaml:"watcher" json:"watcher"`
	Buffer               Buffer            `yaml:"-" json:"buffer"`
	ErrorOutputPattern   string            `yaml:"errorOutputPattern,omitempty" json:"errorOutputPattern,omitempty"`
}

// Command options
type Command struct {
	Type    string `yaml:"type" json:"type"`
	Command string `yaml:"command" json:"command"`
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
	Global  bool   `yaml:"global,omitempty" json:"global,omitempty"`
	Output  bool   `yaml:"output,omitempty" json:"output,omitempty"`
}

// BufferOut is used for exchange information between "realize cli" and "web realize"
type BufferOut struct {
	Time   time.Time `json:"time"`
	Text   string    `json:"text"`
	Path   string    `json:"path"`
	Type   string    `json:"type"`
	Stream string    `json:"stream"`
	Errors []string  `json:"errors"`
}

// Watch the project
func (p *Project) watch() {
	p.watcher, _ = Watcher()
	stop, exit := make(chan bool), make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	// before global commands
	p.cmd(stop, "before", true)
	// indexing files and dirs
	for _, dir := range p.Watcher.Paths {
		base := filepath.Join(p.base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, p.walk); err == nil {
				p.tool(stop, base)
			}
		} else {
			p.err(err)
		}
	}
	// indexing done, files and folders
	msg = fmt.Sprintln(p.pname(p.Name, 1), ":", blue.bold("Watching"), magenta.bold(p.files), "file/s", magenta.bold(p.folders), "folder/s")
	out = BufferOut{Time: time.Now(), Text: "Watching " + strconv.FormatInt(p.files, 10) + " files/s " + strconv.FormatInt(p.folders, 10) + " folder/s"}
	p.stamp("log", out, msg, "")
	// start
	go p.routines(stop, p.watcher, "")
	//is watching
L:
	for {
		select {
		case event := <-p.watcher.Events():
			if time.Now().Truncate(time.Second).After(p.lastTime) || event.Name != p.lastFile {
				switch event.Op {
				case fsnotify.Chmod:
				case fsnotify.Remove:
					ext := ext(event.Name)
					if !strings.Contains(ext, "_") && ext != "" {
						close(stop)
						stop = make(chan bool)
						p.changed(event, stop) // stop
					}
					p.watcher.Remove(event.Name)
				default:
					file, err := os.Stat(event.Name)
					if err != nil {
						continue
					}
					if file.IsDir() {
						if time.Now().Truncate(time.Second).After(p.lastTime) {
							filepath.Walk(event.Name, p.walk)
						}
					} else if file.Size() > 0 {
						if p.parent.Settings.Recovery {
							log.Println(event)
						}
						ext := ext(event.Name)
						if (!strings.Contains(ext, "_") || !strings.Contains(ext, ".")) && array(ext, p.Watcher.Exts) {
							// change watched
							close(stop)
							stop = make(chan bool)
							p.changed(event, stop)
						}
						p.lastTime = time.Now().Truncate(time.Second)
						p.lastFile = event.Name
					}
				}
			}
		case err := <-p.watcher.Errors():
			p.err(err)
		case <-exit:
			p.cmd(nil, "after", true)
			break L
		}
	}
	wg.Done()
	return
}

// Error occurred
func (p *Project) err(err error) {
	msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.regular(err.Error()))
	out = BufferOut{Time: time.Now(), Text: err.Error()}
	p.stamp("error", out, msg, "")
}

// Cmd calls the method that execute commands after/before and display the results
func (p *Project) cmd(stop <-chan bool, flag string, global bool) {
	done := make(chan bool)
	// cmds are scheduled in sequence
	go func() {
		for _, cmd := range p.Watcher.Scripts {
			if strings.ToLower(cmd.Type) == flag && cmd.Global == global {
				err, logs := p.command(stop, cmd)
				if err == "" && logs == "" {
					continue
				}
				msg = fmt.Sprintln(p.pname(p.Name, 5), ":", green.bold("Command"), green.bold("\"")+cmd.Command+green.bold("\""))
				out = BufferOut{Time: time.Now(), Text: cmd.Command, Type: flag}
				if err != "" {
					p.stamp("error", out, msg, "")

					msg = fmt.Sprintln(red.regular(err))
					out = BufferOut{Time: time.Now(), Text: err, Type: flag}
					p.stamp("error", out, "", msg)

				} else if logs != "" && cmd.Output {
					msg = fmt.Sprintln(logs)
					out = BufferOut{Time: time.Now(), Text: logs, Type: flag}
					p.stamp("log", out, "", msg)
				} else {
					p.stamp("log", out, msg, "")
				}
			}
		}
		close(done)
	}()
	for {
		select {
		case <-stop:
			return
		case <-done:
			return
		}
	}
}

// Compile is used for run and display the result of a compiling
func (p *Project) compile(stop <-chan bool, cmd Cmd) error {
	if cmd.Status {
		start := time.Now()
		channel := make(chan Result)
		go func() {
			log.Println(p.pname(p.Name, 1), ":", cmd.startTxt)
			stream, err := p.goCompile(stop, cmd.method, cmd.Args)
			if stream != msgStop {
				channel <- Result{stream, err}
			}
		}()
		select {
		case r := <-channel:
			if r.err != nil {
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.bold(cmd.name), red.regular(r.err.Error()))
				out = BufferOut{Time: time.Now(), Text: r.err.Error(), Type: cmd.name, Stream: r.stream}
				p.stamp("error", out, msg, r.stream)
			} else {
				msg = fmt.Sprintln(p.pname(p.Name, 5), ":", green.regular(cmd.endTxt), "in", magenta.regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
				out = BufferOut{Time: time.Now(), Text: cmd.name + " in " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
				p.stamp("log", out, msg, r.stream)
			}
			return r.err
		case <-stop:
			return nil
		}
	}
	return nil
}

// Defines the colors scheme for the project name
func (p *Project) pname(name string, color int) string {
	switch color {
	case 1:
		name = yellow.regular("[") + strings.ToUpper(name) + yellow.regular("]")
		break
	case 2:
		name = yellow.regular("[") + red.bold(strings.ToUpper(name)) + yellow.regular("]")
		break
	case 3:
		name = yellow.regular("[") + blue.bold(strings.ToUpper(name)) + yellow.regular("]")
		break
	case 4:
		name = yellow.regular("[") + magenta.bold(strings.ToUpper(name)) + yellow.regular("]")
		break
	case 5:
		name = yellow.regular("[") + green.bold(strings.ToUpper(name)) + yellow.regular("]")
		break
	}
	return name
}

//  Tool logs the result of a go command
func (p *Project) tool(stop <-chan bool, path string) error {
	if len(path) > 0 {
		done := make(chan bool)
		result := make(chan tool)
		go func() {
			var wg sync.WaitGroup
			wg.Add(len(p.tools))
			for _, element := range p.tools {
				// no need a sequence, these commands can be asynchronous
				go p.goTool(&wg, stop, result, path, element)
			}
			wg.Wait()
			close(done)
		}()
	loop:
		for {
			select {
			case tool := <-result:
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.bold(tool.name), red.regular("there are some errors in"), ":", magenta.bold(path))
				buff := BufferOut{Time: time.Now(), Text: "there are some errors in", Path: path, Type: tool.name, Stream: tool.err}
				p.stamp("error", buff, msg, tool.err)
			case <-done:
				break loop
			case <-stop:
				break loop
			}
		}
	}
	return nil
}

// Changed detect a file/directory change
func (p *Project) changed(event fsnotify.Event, stop chan bool) {
	e := ext(event.Name)
	if e == "" {
		e = "DIR"
	}
	msg = fmt.Sprintln(p.pname(p.Name, 4), ":", magenta.bold(strings.ToUpper(e)), "changed", magenta.bold(event.Name))
	out = BufferOut{Time: time.Now(), Text: ext(event.Name) + " changed " + event.Name}
	p.stamp("log", out, msg, "")
	//stop running process
	go p.routines(stop, p.watcher, event.Name)
}

// Watch the files tree of a project
func (p *Project) walk(path string, info os.FileInfo, err error) error {
	for _, v := range p.Watcher.Ignore {
		if strings.Contains(path, filepath.Join(p.base, v)) {
			return nil
		}
	}
	if !strings.Contains(path, "/.") && !strings.HasPrefix(path, ".") && (info.IsDir() || array(ext(path), p.Watcher.Exts)) {
		result := p.watcher.Walk(path, p.init)
		if result != "" {
			if info.IsDir() {
				p.folders++
			} else {
				p.files++
			}
			if p.Watcher.Preview {
				log.Println(p.pname(p.Name, 1), ":", path)
			}
		}
	}
	return nil
}

// Print on files, cli, ws
func (p *Project) stamp(t string, o BufferOut, msg string, stream string) {
	switch t {
	case "out":
		p.Buffer.StdOut = append(p.Buffer.StdOut, o)
		if p.Files.Outputs.Status {
			f := p.create(p.base, p.Files.Outputs.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.fatal(err, "")
			}
		}
	case "log":
		p.Buffer.StdLog = append(p.Buffer.StdLog, o)
		if p.Files.Logs.Status {
			f := p.create(p.base, p.Files.Logs.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.fatal(err, "")
			}
		}
	case "error":
		p.Buffer.StdErr = append(p.Buffer.StdErr, o)
		if p.Files.Errors.Status {
			f := p.create(p.base, p.Files.Errors.Name)
			t := time.Now()
			s := []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n"}
			if stream != "" {
				s = []string{t.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Type, o.Text, o.Path, "\r\n", stream}
			}
			if _, err := f.WriteString(strings.Join(s, " ")); err != nil {
				p.fatal(err, "")
			}
		}
	}
	if msg != "" {
		log.Print(msg)
	}
	if stream != "" {
		fmt.Fprint(output, stream)
	}
	go func() {
		p.parent.sync <- "sync"
	}()
}

// Routines launches the toolchain run, build, install
func (p *Project) routines(stop <-chan bool, watcher FileWatcher, path string) {
	var done bool
	var install, build error
	go func() {
		for {
			select {
			case <-stop:
				done = true
				return
			}
		}
	}()
	if !done {
		// before command
		p.cmd(stop, "before", false)
	}
	if !done {
		// Go supported tools
		p.tool(stop, path)
		// Prevent fake events on polling startup
		p.init = true
	}
	if !done {
		install = p.compile(stop, p.Cmds.Install)
	}
	if !done {
		build = p.compile(stop, p.Cmds.Build)
	}
	if !done && (install == nil || build == nil) {
		if p.Cmds.Run {
			start := time.Now()
			runner := make(chan bool, 1)
			go func() {
				log.Println(p.pname(p.Name, 1), ":", "Running..")
				p.goRun(stop, runner)
			}()
			select {
			case <-runner:
				msg = fmt.Sprintln(p.pname(p.Name, 5), ":", green.regular("Started"), "in", magenta.regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
				out = BufferOut{Time: time.Now(), Text: "Started in " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
				p.stamp("log", out, msg, "")
			case <-stop:
				return
			}
		}
	}
	if !done {
		p.cmd(stop, "after", false)
	}
}
