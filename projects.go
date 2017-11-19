package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	msg string
	out BufferOut
)

// Watch info
type Watch struct {
	Paths   []string  `yaml:"paths" json:"paths"`
	Exts    []string  `yaml:"extensions" json:"extensions"`
	Ignore  []string  `yaml:"ignored_paths,omitempty" json:"ignored_paths,omitempty"`
	Scripts []Command `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// Project info
type Project struct {
	parent             *Realize
	watcher            FileWatcher
	init               bool
	files              int64
	folders            int64
	name               string
	lastFile           string
	paths              []string
	lastTime           time.Time
	Name               string            `yaml:"name" json:"name"`
	Path               string            `yaml:"path" json:"path"`
	Environment        map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Tools              Tools             `yaml:"commands" json:"commands"`
	Args               []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Watcher            Watch             `yaml:"watcher" json:"watcher"`
	Buffer             Buffer            `yaml:"-" json:"buffer"`
	ErrorOutputPattern string            `yaml:"errorOutputPattern,omitempty" json:"errorOutputPattern,omitempty"`
}

// Response exec
type Response struct {
	Name string
	Out  string
	Err  error
}

// Buffer define an array buffer for each log files
type Buffer struct {
	StdOut []BufferOut `json:"stdOut"`
	StdLog []BufferOut `json:"stdLog"`
	StdErr []BufferOut `json:"stdErr"`
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

// Setup a project
func (p *Project) Setup() {
	// get base path
	p.name = filepath.Base(p.Path)
	// set env const
	for key, item := range p.Environment {
		if err := os.Setenv(key, item); err != nil {
			p.Buffer.StdErr = append(p.Buffer.StdErr, BufferOut{Time: time.Now(), Text: err.Error(), Type: "Env error", Stream: ""})
		}
	}
	p.Tools.Setup()
}

// Watch a project
func (p *Project) Watch(exit chan os.Signal) {
	var err error
	stop := make(chan bool)
	// init a new watcher
	p.watcher, err = Watcher(p.parent.Settings.Legacy.Force, p.parent.Settings.Legacy.Interval)
	if err != nil {
		log.Fatal(err)
	}
	// global commands before
	p.cmd(stop, "before", true)
	// indexing files and dirs
	for _, dir := range p.Watcher.Paths {
		base, _ := filepath.Abs(p.Path)
		base = filepath.Join(base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, p.walk); err == nil {
				p.tools(stop, base)
			}
		} else {
			p.err(err)
		}
	}
	// start message
	msg = fmt.Sprintln(p.pname(p.Name, 1), ":", blue.bold("Watching"), magenta.bold(p.files), "file/s", magenta.bold(p.folders), "folder/s")
	out = BufferOut{Time: time.Now(), Text: "Watching " + strconv.FormatInt(p.files, 10) + " files/s " + strconv.FormatInt(p.folders, 10) + " folder/s"}
	p.stamp("log", out, msg, "")
	// start watcher
	go p.Reload(p.watcher, "", stop)
L:
	for {
		select {
		case event := <-p.watcher.Events():
			if time.Now().Truncate(time.Second).After(p.lastTime) || event.Name != p.lastFile {
				// event time
				eventTime := time.Now()
				// file extension
				ext := ext(event.Name)
				if ext == "" {
					ext = "DIR"
				}
				// change message
				msg = fmt.Sprintln(p.pname(p.Name, 4), ":", magenta.bold(strings.ToUpper(ext)), "changed", magenta.bold(event.Name))
				out = BufferOut{Time: time.Now(), Text: ext + " changed " + event.Name}
				// switch event type
				switch event.Op {
				case fsnotify.Chmod:
				case fsnotify.Remove:
					p.watcher.Remove(event.Name)
					if !strings.Contains(ext, "_") && !strings.Contains(ext, ".") && array(ext, p.Watcher.Exts) {
						close(stop)
						stop = make(chan bool)
						p.stamp("log", out, msg, "")
						go p.Reload(p.watcher, "", stop)
					}
				default:
					file, err := os.Stat(event.Name)
					if err != nil {
						continue
					}
					if file.IsDir() {
						filepath.Walk(event.Name, p.walk)
					} else if file.Size() > 0 {
						if !strings.Contains(ext, "_") && !strings.Contains(ext, ".") && array(ext, p.Watcher.Exts) {
							// change watched
							// check if a file is still writing #119
							if event.Op != fsnotify.Write || (eventTime.Truncate(time.Millisecond).After(file.ModTime().Truncate(time.Millisecond)) || event.Name != p.lastFile) {
								close(stop)
								stop = make(chan bool)
								// stop and start again
								p.stamp("log", out, msg, "")
								go p.Reload(p.watcher, event.Name, stop)
							}
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
}

// Reload launches the toolchain run, build, install
func (p *Project) Reload(watcher FileWatcher, path string, stop <-chan bool) {
	var done bool
	var install, build Response
	go func() {
		for {
			select {
			case <-stop:
				done = true
				return
			}
		}
	}()
	if done {
		return
	}
	// before command
	p.cmd(stop, "before", false)
	if done {
		return
	}
	// Go supported tools
	p.tools(stop, path)
	// Prevent fake events on polling startup
	p.init = true
	// prevent errors using realize without config with only run flag
	if p.Tools.Run && !p.Tools.Install.Status && !p.Tools.Build.Status {
		p.Tools.Install.Status = true
	}
	if done {
		return
	}
	if p.Tools.Install.Status {
		msg = fmt.Sprintln(p.pname(p.Name, 1), ":", green.regular(p.Tools.Install.name), "started")
		out = BufferOut{Time: time.Now(), Text: p.Tools.Install.name + " started"}
		p.stamp("log", out, msg, "")
		start := time.Now()
		install = p.Tools.Install.Compile(p.Path, stop)
		install.printAfter(start, p)
	}
	if done {
		return
	}
	if p.Tools.Build.Status {
		msg = fmt.Sprintln(p.pname(p.Name, 1), ":", green.regular(p.Tools.Build.name), "started")
		out = BufferOut{Time: time.Now(), Text: p.Tools.Build.name + " started"}
		p.stamp("log", out, msg, "")
		start := time.Now()
		build = p.Tools.Build.Compile(p.Path, stop)
		build.printAfter(start, p)
	}
	if done {
		return
	}
	if install.Err == nil && build.Err == nil && p.Tools.Run {
		var start time.Time
		result := make(chan Response)
		go func() {
			select {
			case r := <-result:
				if r.Err != nil {
					msg := fmt.Sprintln(p.pname(p.Name, 2), ":", red.regular(r.Err))
					out := BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: "Go Run"}
					p.stamp("error", out, msg, "")
				}
				if r.Out != "" {
					msg := fmt.Sprintln(p.pname(p.Name, 3), ":", blue.regular(r.Out))
					out := BufferOut{Time: time.Now(), Text: r.Out, Type: "Go Run"}
					p.stamp("out", out, msg, "")
				}
			}
		}()
		go func() {
			log.Println(p.pname(p.Name, 1), ":", "Running..")
			start = time.Now()
			p.Run(p.Path, stop)
		}()
	}
	if done {
		return
	}
	p.cmd(stop, "after", false)
}

// Run a project
func (p *Project) Run(path string, stop <-chan bool) (response chan Response) {
	var args []string
	var build *exec.Cmd
	var r Response
	defer func() {
		if err := build.Process.Kill(); err != nil {
			r.Err = err
		}
	}()

	// custom error pattern
	isErrorText := func(string) bool {
		return false
	}
	errRegexp, err := regexp.Compile(p.ErrorOutputPattern)
	if err != nil {
		r.Err = err
		response <- r
		r.Err = nil
	} else {
		isErrorText = func(t string) bool {
			return errRegexp.MatchString(t)
		}
	}

	// add additional arguments
	for _, arg := range p.Args {
		a := strings.FieldsFunc(arg, func(i rune) bool {
			return i == '"' || i == '=' || i == '\''
		})
		args = append(args, a...)
	}
	gobin := os.Getenv("GOBIN")
	dirPath := filepath.Base(path)
	if path == "." {
		dirPath = filepath.Base(wdir())
	}
	path = filepath.Join(gobin, dirPath)
	if _, err := os.Stat(path); err == nil {
		build = exec.Command(path, args...)
	} else if _, err := os.Stat(path + RExtWin); err == nil {
		build = exec.Command(path+RExtWin, args...)
	} else {
		if _, err = os.Stat(path); err == nil {
			build = exec.Command(path, args...)
		} else if _, err = os.Stat(path + RExtWin); err == nil {
			build = exec.Command(path+RExtWin, args...)
		} else {
			r.Err = errors.New("project not found")
			return
		}
	}
	// scan project stream
	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()
	if err != nil {
		r.Err = err
		return
	}
	if err := build.Start(); err != nil {
		r.Err = err
		return
	}
	execOutput, execError := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOutput, stopError := make(chan bool, 1), make(chan bool, 1)
	scanner := func(stop chan bool, output *bufio.Scanner, isError bool) {
		for output.Scan() {
			text := output.Text()
			if isError && !isErrorText(text) {
				r.Err = errors.New(text)
				response <- r
				r.Err = nil
			} else {
				r.Out = text
				response <- r
				r.Out = ""
			}
		}
		close(stop)
	}
	go scanner(stopOutput, execOutput, false)
	go scanner(stopError, execError, true)
	for {
		select {
		case <-stop:
			return
		case <-stopOutput:
			return
		case <-stopError:
			return
		}
	}
}

// Error occurred
func (p *Project) err(err error) {
	msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.regular(err.Error()))
	out = BufferOut{Time: time.Now(), Text: err.Error()}
	p.stamp("error", out, msg, "")
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
func (p *Project) tools(stop <-chan bool, path string) {
	if len(path) > 0 {
		done := make(chan bool)
		result := make(chan Response)
		v := reflect.ValueOf(p.Tools)
		go func() {
			for i := 0; i < v.NumField()-1; i++ {
				tool := v.Field(i).Interface().(Tool)
				if tool.Status && tool.isTool {
					result <- tool.Exec(path, stop)
				}
			}
			close(done)
		}()
		for {
			select {
			case <-done:
				return
			case <-stop:
				return
			case r := <-result:
				if r.Err != nil {
					msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.bold(r.Name), red.regular("there are some errors in"), ":", magenta.bold(path))
					buff := BufferOut{Time: time.Now(), Text: "there are some errors in", Path: path, Type: r.Name, Stream: r.Err.Error()}
					p.stamp("error", buff, msg, r.Err.Error())
				} else if r.Out != "" {
					msg = fmt.Sprintln(p.pname(p.Name, 3), ":", red.bold(r.Name), red.regular("outputs"), ":", blue.bold(path))
					buff := BufferOut{Time: time.Now(), Text: "outputs", Path: path, Type: r.Name, Stream: r.Out}
					p.stamp("out", buff, msg, r.Out)
				}
			}
		}
	}
}

// Cmd after/before
func (p *Project) cmd(stop <-chan bool, flag string, global bool) {
	done := make(chan bool)
	result := make(chan Response)
	// commands sequence
	go func() {
		for _, cmd := range p.Watcher.Scripts {
			if strings.ToLower(cmd.Type) == flag && cmd.Global == global {
				result <- cmd.Exec(p.Path, stop)
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
		case r := <-result:
			msg = fmt.Sprintln(p.pname(p.Name, 5), ":", green.bold("Command"), green.bold("\"")+r.Name+green.bold("\""))
			out = BufferOut{Time: time.Now(), Text: r.Name, Type: flag}
			if r.Err != nil {
				p.stamp("error", out, msg, "")
				out = BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: flag}
				p.stamp("error", out, "", fmt.Sprintln(red.regular(r.Err.Error())))
			}
			if r.Out != "" {
				out = BufferOut{Time: time.Now(), Text: r.Out, Type: flag}
				p.stamp("log", out, "", fmt.Sprintln(r.Out))
			} else {
				p.stamp("log", out, msg, "")
			}
		}
	}
}

// Watch the files tree of a project
func (p *Project) walk(path string, info os.FileInfo, err error) error {
	for _, v := range p.Watcher.Ignore {
		s := append([]string{p.Path}, strings.Split(v, string(os.PathSeparator))...)
		if strings.Contains(path, filepath.Join(s...)) {
			return nil
		}
	}
	if !strings.HasPrefix(path, ".") && (info.IsDir() || array(ext(path), p.Watcher.Exts)) {
		result := p.watcher.Walk(path, p.init)
		if result != "" {
			if info.IsDir() {
				p.folders++
			} else {
				p.files++
			}
		}
	}
	return nil
}

// Print on files, cli, ws
func (p *Project) stamp(t string, o BufferOut, msg string, stream string) {
	time := time.Now()
	content := []string{time.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n", stream}
	switch t {
	case "out":
		p.Buffer.StdOut = append(p.Buffer.StdOut, o)
		if p.parent.Settings.Files.Outputs.Status {
			f := p.parent.Settings.Create(p.Path, p.parent.Settings.Files.Outputs.Name)
			if _, err := f.WriteString(strings.Join(content, " ")); err != nil {
				p.parent.Settings.Fatal(err, "")
			}
		}
	case "log":
		p.Buffer.StdLog = append(p.Buffer.StdLog, o)
		if p.parent.Settings.Files.Logs.Status {
			f := p.parent.Settings.Create(p.Path, p.parent.Settings.Files.Logs.Name)
			if _, err := f.WriteString(strings.Join(content, " ")); err != nil {
				p.parent.Settings.Fatal(err, "")
			}
		}
	case "error":
		p.Buffer.StdErr = append(p.Buffer.StdErr, o)
		if p.parent.Settings.Files.Errors.Status {
			f := p.parent.Settings.Create(p.Path, p.parent.Settings.Files.Errors.Name)
			if _, err := f.WriteString(strings.Join(content, " ")); err != nil {
				p.parent.Settings.Fatal(err, "")
			}
		}
	}
	if msg != "" {
		log.Print(msg)
	}
	if stream != "" {
		fmt.Fprint(output, stream)
	}
}

func (r *Response) printAfter(start time.Time, p *Project) {
	if r.Err != nil {
		msg = fmt.Sprintln(p.pname(p.Name, 2), ":", red.bold(r.Name), red.regular(r.Err.Error()))
		out = BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: r.Name, Stream: r.Out}
		p.stamp("error", out, msg, r.Out)
	} else {
		msg = fmt.Sprintln(p.pname(p.Name, 5), ":", green.bold(r.Name), "completed in", magenta.regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
		out = BufferOut{Time: time.Now(), Text: r.Name + " in " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
		p.stamp("log", out, msg, r.Out)
	}
}
