package realize

import (
	"bufio"
	"bytes"
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
	"sync"
	"time"
)

var (
	msg string
	out BufferOut
)

// Watch info
type Watch struct {
	Exts    []string  `yaml:"extensions" json:"extensions"`
	Paths   []string  `yaml:"paths" json:"paths"`
	Scripts []Command `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	Hidden  bool      `yaml:"hidden,omitempty" json:"hidden,omitempty"`
	Ignore  Ignore    `yaml:"ignore,omitempty" json:"ignore,omitempty"`
}

type Ignore struct{
	Exts   []string  `yaml:"exts,omitempty" json:"exts,omitempty"`
	Paths  []string  `yaml:"paths,omitempty" json:"paths,omitempty"`
}

// Command fields
type Command struct {
	Cmd    string `yaml:"command" json:"command"`
	Type   string `yaml:"type" json:"type"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Global bool   `yaml:"global,omitempty" json:"global,omitempty"`
	Output bool   `yaml:"output,omitempty" json:"output,omitempty"`
}

// Project info
type Project struct {
	parent             *Realize
	watcher            FileWatcher
	stop               chan bool
	exit               chan os.Signal
	paths              []string
	last			   last
	files              int64
	folders            int64
	init               bool
	Name               string            `yaml:"name" json:"name"`
	Path               string            `yaml:"path" json:"path"`
	Env        	   map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Args               []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Tools              Tools             `yaml:"commands" json:"commands"`
	Watcher            Watch             `yaml:"watcher" json:"watcher"`
	Buffer             Buffer            `yaml:"-" json:"buffer"`
	ErrPattern string            `yaml:"pattern,omitempty" json:"pattern,omitempty"`
}

// Last is used to save info about last file changed
type last struct {
	file string
	time time.Time
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

// After stop watcher
func (p *Project) After() {
	if p.parent.After != nil {
		p.parent.After(Context{Project: p})
		return
	}
	p.cmd(nil, "after", true)
}

// Before start watcher
func (p *Project) Before() {
	if p.parent.Before != nil {
		p.parent.Before(Context{Project: p})
		return
	}
	// setup go tools
	p.Tools.Setup()
	// set env const
	for key, item := range p.Env {
		if err := os.Setenv(key, item); err != nil {
			p.Buffer.StdErr = append(p.Buffer.StdErr, BufferOut{Time: time.Now(), Text: err.Error(), Type: "Env error", Stream: ""})
		}
	}
	// global commands before
	p.cmd(p.stop, "before", true)
	// indexing files and dirs
	for _, dir := range p.Watcher.Paths {
		base, _ := filepath.Abs(p.Path)
		base = filepath.Join(base, dir)
		if _, err := os.Stat(base); err == nil {
			if err := filepath.Walk(base, p.walk); err != nil {
				p.Err(err)
			}
		}
	}
	// start message
	msg = fmt.Sprintln(p.pname(p.Name, 1), ":", Blue.Bold("Watching"), Magenta.Bold(p.files), "file/s", Magenta.Bold(p.folders), "folder/s")
	out = BufferOut{Time: time.Now(), Text: "Watching " + strconv.FormatInt(p.files, 10) + " files/s " + strconv.FormatInt(p.folders, 10) + " folder/s"}
	p.stamp("log", out, msg, "")
}

// Err occurred
func (p *Project) Err(err error) {
	if p.parent.Err != nil {
		p.parent.Err(Context{Project: p})
		return
	}
	if err != nil {
		msg = fmt.Sprintln(p.pname(p.Name, 2), ":", Red.Regular(err.Error()))
		out = BufferOut{Time: time.Now(), Text: err.Error()}
		p.stamp("error", out, msg, "")
	}
}

// Change event message
func (p *Project) Change(event fsnotify.Event) {
	if p.parent.Change != nil {
		p.parent.Change(Context{Project: p, Event: event})
		return
	}
	// file extension
	ext := ext(event.Name)
	if ext == "" {
		ext = "DIR"
	}
	// change message
	msg = fmt.Sprintln(p.pname(p.Name, 4), ":", Magenta.Bold(strings.ToUpper(ext)), "changed", Magenta.Bold(event.Name))
	out = BufferOut{Time: time.Now(), Text: ext + " changed " + event.Name}
	p.stamp("log", out, msg, "")
}

// Reload launches the toolchain run, build, install
func (p *Project) Reload(path string, stop <-chan bool) {
	if p.parent.Reload != nil {
		p.parent.Reload(Context{Project: p, Watcher: p.watcher, Path: path, Stop: stop})
		return
	}
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
	if len(path) > 0 {
		fi, err := os.Stat(path)
		if filepath.Ext(path) == "" {
			fi, err = os.Stat(path)
		}
		if err != nil {
			p.Err(err)
		}
		p.tools(stop, path, fi)
	}
	// Prevent fake events on polling startup
	p.init = true
	// prevent errors using realize without config with only run flag
	if p.Tools.Run.Status && !p.Tools.Install.Status && !p.Tools.Build.Status {
		p.Tools.Install.Status = true
	}
	if done {
		return
	}
	if p.Tools.Install.Status {
		msg = fmt.Sprintln(p.pname(p.Name, 1), ":", Green.Regular(p.Tools.Install.name), "started")
		out = BufferOut{Time: time.Now(), Text: p.Tools.Install.name + " started"}
		p.stamp("log", out, msg, "")
		start := time.Now()
		install = p.Tools.Install.Compile(p.Path, stop)
		install.print(start, p)
	}
	if done {
		return
	}
	if p.Tools.Build.Status {
		msg = fmt.Sprintln(p.pname(p.Name, 1), ":", Green.Regular(p.Tools.Build.name), "started")
		out = BufferOut{Time: time.Now(), Text: p.Tools.Build.name + " started"}
		p.stamp("log", out, msg, "")
		start := time.Now()
		build = p.Tools.Build.Compile(p.Path, stop)
		build.print(start, p)
	}
	if done {
		return
	}
	if install.Err == nil && build.Err == nil && p.Tools.Run.Status {
		result := make(chan Response)
		go func() {
			for {
				select {
				case <-stop:
					return
				case r := <-result:
					if r.Err != nil {
						msg := fmt.Sprintln(p.pname(p.Name, 2), ":", Red.Regular(r.Err))
						out := BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: "Go Run"}
						p.stamp("error", out, msg, "")
					}
					if r.Out != "" {
						msg := fmt.Sprintln(p.pname(p.Name, 3), ":", Blue.Regular(r.Out))
						out := BufferOut{Time: time.Now(), Text: r.Out, Type: "Go Run"}
						p.stamp("out", out, msg, "")
					}
				}
			}
		}()
		go func() {
			log.Println(p.pname(p.Name, 1), ":", "Running..")
			err := p.run(p.Path, result, stop)
			if err != nil {
				msg := fmt.Sprintln(p.pname(p.Name, 2), ":", Red.Regular(err))
				out := BufferOut{Time: time.Now(), Text: err.Error(), Type: "Go Run"}
				p.stamp("error", out, msg, "")
			}
		}()
	}
	if done {
		return
	}
	p.cmd(stop, "after", false)
}

// Watch a project
func (p *Project) Watch(wg *sync.WaitGroup) {
	var err error
	// change channel
	p.stop = make(chan bool)
	// init a new watcher
	p.watcher, err = NewFileWatcher(p.parent.Settings.Legacy)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		close(p.stop)
		p.watcher.Close()
	}()
	// before start checks
	p.Before()
	// start watcher
	go p.Reload("", p.stop)
L:
	for {
		select {
		case event := <-p.watcher.Events():
			if p.parent.Settings.Recovery.Events {
				log.Println("File:", event.Name, "LastFile:", p.last.file, "Time:", time.Now(), "LastTime:", p.last.time)
			}
			if time.Now().Truncate(time.Second).After(p.last.time) {
				// switch event type
				switch event.Op {
				case fsnotify.Chmod:
				case fsnotify.Remove:
					p.watcher.Remove(event.Name)
					if p.Validate(event.Name, false) && ext(event.Name) != "" {
						// stop and restart
						close(p.stop)
						p.stop = make(chan bool)
						p.Change(event)
						go p.Reload("", p.stop)
					}
				default:
					if p.Validate(event.Name, true) {
						fi, err := os.Stat(event.Name)
						if err != nil {
							continue
						}
						if fi.IsDir() {
							filepath.Walk(event.Name, p.walk)
						} else {
							// stop and restart
							close(p.stop)
							p.stop = make(chan bool)
							p.Change(event)
							go p.Reload(event.Name, p.stop)
							p.last.time = time.Now().Truncate(time.Second)
							p.last.file = event.Name
						}
					}
				}
			}
		case err := <-p.watcher.Errors():
			p.Err(err)
		case <-p.exit:
			p.After()
			break L
		}
	}
	wg.Done()
}

// Validate a file path
func (p *Project) Validate(path string, fcheck bool) bool {
	if len(path) <= 0 {
		return false
	}
	// check if skip hidden
	if p.Watcher.Hidden && isHidden(path) {
		return false
	}
	// check for a valid ext or path
	if e := ext(path); e != "" {
		if len(p.Watcher.Exts) == 0{
			return false
		}
		// check ignored
		for _, v := range p.Watcher.Ignore.Exts {
			if v == e {
				return false
			}
		}
		// supported extensions
		for index, v := range p.Watcher.Exts{
			if e == v {
				break
			}
			if index == len(p.Watcher.Exts)-1{
				return false
			}
		}
	}
	separator := string(os.PathSeparator)
	// supported paths
	for _, v := range p.Watcher.Ignore.Paths {
		s := append([]string{p.Path}, strings.Split(v, separator)...)
		abs, _ := filepath.Abs(filepath.Join(s...))
		if path == abs || strings.HasPrefix(path, abs+separator) {
			return false
		}
	}
	// file check
	if fcheck {
		fi, err := os.Stat(path)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() && ext(path) == "" || fi.Size() <= 0{
			return false
		}
	}
	return true

}

// Defines the colors scheme for the project name
func (p *Project) pname(name string, color int) string {
	switch color {
	case 1:
		name = Yellow.Regular("[") + strings.ToUpper(name) + Yellow.Regular("]")
		break
	case 2:
		name = Yellow.Regular("[") + Red.Bold(strings.ToUpper(name)) + Yellow.Regular("]")
		break
	case 3:
		name = Yellow.Regular("[") + Blue.Bold(strings.ToUpper(name)) + Yellow.Regular("]")
		break
	case 4:
		name = Yellow.Regular("[") + Magenta.Bold(strings.ToUpper(name)) + Yellow.Regular("]")
		break
	case 5:
		name = Yellow.Regular("[") + Green.Bold(strings.ToUpper(name)) + Yellow.Regular("]")
		break
	}
	return name
}

//  Tool logs the result of a go command
func (p *Project) tools(stop <-chan bool, path string, fi os.FileInfo) {
	done := make(chan bool)
	result := make(chan Response)
	v := reflect.ValueOf(p.Tools)
	go func() {
		for i := 0; i < v.NumField()-1; i++ {
			tool := v.Field(i).Interface().(Tool)
			tool.parent = p
			if tool.Status && tool.isTool {
				if fi.IsDir() {
					if tool.dir {
						result <- tool.Exec(path, stop)
					}
				} else if !tool.dir {
					result <- tool.Exec(path, stop)
				}
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
				if fi.IsDir() {
					path, _ = filepath.Abs(fi.Name())
				}
				msg = fmt.Sprintln(p.pname(p.Name, 2), ":", Red.Bold(r.Name), Red.Regular("there are some errors in"), ":", Magenta.Bold(path))
				buff := BufferOut{Time: time.Now(), Text: "there are some errors in", Path: path, Type: r.Name, Stream: r.Err.Error()}
				p.stamp("error", buff, msg, r.Err.Error())
			} else if r.Out != "" {
				msg = fmt.Sprintln(p.pname(p.Name, 3), ":", Red.Bold(r.Name), Red.Regular("outputs"), ":", Blue.Bold(path))
				buff := BufferOut{Time: time.Now(), Text: "outputs", Path: path, Type: r.Name, Stream: r.Out}
				p.stamp("out", buff, msg, r.Out)
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
				result <- cmd.exec(p.Path, stop)
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
			 msg = fmt.Sprintln(p.pname(p.Name, 5), ":", Green.Bold("Command"), Green.Bold("\"")+r.Name+Green.Bold("\""))
			 if r.Err != nil {
				out = BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: flag}
				p.stamp("error", out, msg, fmt.Sprint(Red.Regular(r.Err.Error())))
			} else {
				out = BufferOut{Time: time.Now(), Text: r.Out, Type: flag}
				p.stamp("log", out, msg, fmt.Sprint(r.Out))
			}
		}
	}
}

// Watch the files tree of a project
func (p *Project) walk(path string, info os.FileInfo, err error) error {
	if p.Validate(path, true) {
		result := p.watcher.Walk(path, p.init)
		if result != "" {
			if p.parent.Settings.Recovery.Index {
				log.Println("Indexing", path)
			}
			p.tools(p.stop, path, info)
			if info.IsDir() {
				// tools dir
				p.folders++
			} else {
				// tools files
				p.files++
			}
		}
	}
	return nil
}

// Print on files, cli, ws
func (p *Project) stamp(t string, o BufferOut, msg string, stream string) {
	ctime := time.Now()
	content := []string{ctime.Format("2006-01-02 15:04:05"), strings.ToUpper(p.Name), ":", o.Text, "\r\n", stream}
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
		fmt.Fprintln(Output, stream)
	}
	go func() {
		p.parent.Sync <- "sync"
	}()
}

// Run a project
func (p *Project) run(path string, stream chan Response, stop <-chan bool) (err error) {
	var args []string
	var build *exec.Cmd
	var r Response
	defer func() {
		// https://github.com/golang/go/issues/5615
		// https://github.com/golang/go/issues/6720
		if build != nil {
			build.Process.Signal(os.Interrupt)
		}
	}()

	// custom error pattern
	isErrorText := func(string) bool {
		return false
	}
	errRegexp, err := regexp.Compile(p.ErrPattern)
	if err != nil {
		r.Err = err
		stream <- r
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
	dirPath := os.Getenv("GOBIN")
	if p.Tools.Run.Dir != "" {
		dirPath, _ = filepath.Abs(p.Tools.Run.Dir)
	}
	name := filepath.Base(path)
	if path == "." && p.Tools.Run.Dir == "" {
		name = filepath.Base(Wdir())
	} else if p.Tools.Run.Dir != "" {
		name = filepath.Base(dirPath)
	}
	path = filepath.Join(dirPath, name)
	if p.Tools.Run.Method != "" {
        path = p.Tools.Run.Method
	}
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
			return errors.New("project not found")
		}
	}
	// scan project stream
	stdout, err := build.StdoutPipe()
	stderr, err := build.StderrPipe()
	if err != nil {
		return err
	}
	if err := build.Start(); err != nil {
		return err
	}
	execOutput, execError := bufio.NewScanner(stdout), bufio.NewScanner(stderr)
	stopOutput, stopError := make(chan bool, 1), make(chan bool, 1)
	scanner := func(stop chan bool, output *bufio.Scanner, isError bool) {
		for output.Scan() {
			text := output.Text()
			if isError && !isErrorText(text) {
				r.Err = errors.New(text)
				stream <- r
				r.Err = nil
			} else {
				r.Out = text
				stream <- r
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

// Print with time after
func (r *Response) print(start time.Time, p *Project) {
	if r.Err != nil {
		msg = fmt.Sprintln(p.pname(p.Name, 2), ":", Red.Bold(r.Name), "\n", r.Err.Error())
		out = BufferOut{Time: time.Now(), Text: r.Err.Error(), Type: r.Name, Stream: r.Out}
		p.stamp("error", out, msg, r.Out)
	} else {
		msg = fmt.Sprintln(p.pname(p.Name, 5), ":", Green.Bold(r.Name), "completed in", Magenta.Regular(big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3), " s"))
		out = BufferOut{Time: time.Now(), Text: r.Name + " in " + big.NewFloat(float64(time.Since(start).Seconds())).Text('f', 3) + " s"}
		p.stamp("log", out, msg, r.Out)
	}
}

// Exec an additional command from a defined path if specified
func (c *Command) exec(base string, stop <-chan bool) (response Response) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan error)
	args := strings.Split(strings.Replace(strings.Replace(c.Cmd, "'", "", -1), "\"", "", -1), " ")
	ex := exec.Command(args[0], args[1:]...)
	ex.Dir = base
	// make cmd path
	if c.Path != "" {
		if strings.Contains(c.Path, base) {
			ex.Dir = c.Path
		} else {
			ex.Dir = filepath.Join(base, c.Path)
		}
	}
	ex.Stdout = &stdout
	ex.Stderr = &stderr
	// Start command
	ex.Start()
	go func() { done <- ex.Wait() }()
	// Wait a result
	select {
	case <-stop:
		// Stop running command
		ex.Process.Kill()
	case err := <-done:
		// Command completed
		response.Name = c.Cmd
		response.Out = stdout.String()
		if err != nil {
			response.Err = errors.New(stderr.String() + stdout.String())
		}
	}
	return
}
