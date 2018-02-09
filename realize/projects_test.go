package realize

import (
	"bytes"
	"errors"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProject_After(t *testing.T) /**/ {
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
		Env: map[string]string{
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
	event := fsnotify.Event{Name: "test", Op: fsnotify.Write}
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
	r.Settings.Legacy.Force = false
	r.Settings.Legacy.Interval = 0
	r.Projects[0].watcher, _ = NewFileWatcher(r.Settings.Legacy)
	r.Reload = func(context Context) {
		log.Println(context.Path)
	}
	stop := make(chan bool)
	r.Projects[0].Reload(input, stop)
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Validate(t *testing.T) {
	data := map[string]bool{
		"":                        false,
		"/test/.path/":            true,
		"./test/path/":            true,
		"/test/path/test.html":    false,
		"/test/path/test.go":      false,
		"/test/ignore/test.go":    false,
		"/test/check/notexist.go": false,
		"/test/check/exist.go":    false,
	}
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
		Watcher: Watch{
			Exts: []string{},
			Ignore: Ignore{
				Paths:[]string{"/test/ignore"},
			},
		},
	})
	for i, v := range data {
		result := r.Projects[0].Validate(i, false)
		if  result != v {
			t.Error("Unexpected error", i, "expected", v, result)
		}
	}
}

func TestProject_Watch(t *testing.T) {
	var wg sync.WaitGroup
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
		exit:   make(chan os.Signal, 1),
	})
	go func() {
		time.Sleep(100)
		close(r.Projects[0].exit)
	}()
	wg.Add(1)
	// test before after and file change
	r.Projects[0].Watch(&wg)
	wg.Wait()
}
