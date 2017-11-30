package realize

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"github.com/fsnotify/fsnotify"
	"os/signal"
	"time"
)

func TestProject_After(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	input := "text"
	r.After = func(context Context) {
		log.Println(input)
	}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	r.Projects[0].After()
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Before(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	input := "text"
	r.Before = func(context Context) {
		log.Println(input)
	}
	r.Projects[0].Before()
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}

	r = Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
		Environment: map[string]string{
			input: input,
		},
	})
	r.Projects[0].Before()
	if os.Getenv(input) != input {
		t.Error("Unexpected error expected", input, "instead", os.Getenv(input))
	}
}

func TestProject_Err(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	input := "text"
	r.Err = func(context Context) {
		log.Println(input)
	}
	r.Projects[0].Err(errors.New(input))
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Change(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	r.Change = func(context Context) {
		log.Println(context.Event.Name)
	}
	event := fsnotify.Event{Name:"test",Op:fsnotify.Write}
	r.Projects[0].Change(event)
	if !strings.Contains(buf.String(), event.Name) {
		t.Error("Unexpected error")
	}
}

func TestProject_Reload(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	input := "test/path"
	r.Projects[0].watcher, _ = NewFileWatcher(false,0)
	r.Reload = func(context Context) {
		log.Println(context.Path)
	}
	stop := make(chan bool)
	r.Projects[0].Reload(input,stop)
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Validate(t *testing.T) {
	data := map[string]bool{
		"": false,
		"/test/.path/": false,
		"./test/path/": false,
		"/test/path/test.html": false,
		"/test/path/test.go": false,
		"/test/ignore/test.go": false,
		"/test/check/notexist.go": false,
		"/test/check/exist.go": false,
	}
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
		Watcher: Watch{
			Ignore: []string{"/test/ignore"},
		},
	})
	for i, v := range data {
		if r.Projects[0].Validate(i,true) != v{
			t.Error("Unexpected error",i,"expected",v)
		}
	}
}

func TestProject_Watch(t *testing.T) {
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	r.exit = make(chan os.Signal, 2)
	signal.Notify(r.exit, os.Interrupt)
	go func(){
		time.Sleep(100)
		close(r.exit)
	}()
	r.Projects[0].Watch(r.exit)
}