package core

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestActivityPush(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{
		Settings: Settings{
			Logs: Logs{
				File: true,
			},
		},
	}
	r.Projects = append(r.Projects, Project{Realize: &r})
	r.Projects[0].Push("test", "push")
	expected := fmt.Sprintln("test", "push")
	if buf.String() != expected {
		t.Fatal("Mismatch", buf.String(), expected)
	}
	if _, err := os.Stat(logs); os.IsNotExist(err) {
		t.Fatal("Expected a log file")
	} else {
		os.Remove(logs)
	}
}

func TestActivityRecover(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{
		Settings: Settings{
			Logs: Logs{
				Recovery: true,
			},
		},
	}
	r.Projects = append(r.Projects, Project{Realize: &r})
	r.Projects[0].Recover("test", "recover")
	expected := fmt.Sprintln("test", "recover")
	if buf.String() != expected {
		t.Fatal("Mismatch", buf.String(), expected)
	}
	// switch off recovery
	buf = bytes.Buffer{}
	r.Settings.Logs.Recovery = false
	r.Projects[0].Recover("test", "recover")
	if len(buf.String()) > 0 {
		t.Fatal("Unexpected string", buf.String())
	}

}

func TestActivityWalk(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var watcher FileWatcher
	watcher, err = NewFileWatcher(Polling{Active: false})
	if err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs(dir)
	path = filepath.Join(dir, string(os.PathSeparator))
	model := Project{Ignore: &Ignore{Hidden: true}}
	if err := model.Walk(path, watcher); err != nil {
		t.Fatal(err)
	}
	var countFiles int
	var countFolders int
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		fi, _ := os.Stat(path)
		if fi.IsDir() && filepath.Base(fi.Name())[0:1] == "." {
			return filepath.SkipDir
		} else if fi.IsDir() {
			countFolders++
		} else if e := Ext(fi.Name()); e != "" {
			if !Hidden(fi.Name()) {
				countFiles++
			}
		}
		return nil
	})
	if len(model.Files) != countFiles {
		t.Fatal("Wrong files count", len(model.Files), countFiles)
	}
	if len(model.Folders) != countFolders {
		t.Fatal("Wrong folders count", len(model.Folders), countFolders)
	}
}

func TestActivityExec(t *testing.T) {
	var buf bytes.Buffer
	var wg sync.WaitGroup
	log.SetOutput(&buf)
	wg.Add(1)
	Project := Project{}
	reload := make(chan bool)
	command := Command{Task: "sleep 1"}
	Project.Exec(command, &wg, reload)
	if buf.Len() == 0 {
		t.Fatal("Unexpected error")
	}
}

func TestActivityScan(t *testing.T) {
	var buf bytes.Buffer
	var wg sync.WaitGroup
	log.SetOutput(&buf)
	wg.Add(1)
	realize := Realize{Exit: make(chan bool)}
	Project := Project{
		Realize: &realize,
		Watch: &Watch{
			Path: []string{
				"../**/*.go",
			},
		},
	}
	sequence := Series{
		Tasks: ToInterface([]Command{
			{
				Task: "echo test",
			},
			{
				Task: "sleep 1",
			},
		}),
	}
	Project.Tasks = make([]interface{}, 0)
	Project.Tasks = append(Project.Tasks, sequence)
	// stop scan after 1.5 sec
	go func() {
		time.Sleep(1500 * time.Millisecond)
		realize.Exit <- true
	}()
	Project.Scan(&wg)
	if buf.Len() == 0 {
		t.Fatal("Unexpected error")
	}

}

func TestActivityReload(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	Project := Project{}
	reload := make(chan bool)
	tasks := make([]interface{}, 0)
	parallel := Parallel{
		Tasks: ToInterface([]Command{
			{
				Task: "echo clean super command root test",
			},
			{
				Task: "go fmt",
			},
		}),
	}
	sequence := Series{
		Tasks: ToInterface([]Command{
			{
				Task: "go install",
			},
			{
				Task: "go build",
			},
		}),
	}
	tasks = append(tasks, parallel)
	tasks = append(tasks, sequence)
	Project.Run(reload, tasks...)
	if buf.Len() == 0 {
		t.Fatal("Unexpected error")
	}
}

func TestActivityValidate(t *testing.T) {
	// Test paths
	paths := map[string]bool{
		"/style.go":       true,
		"./handle.go":     true,
		"/settings.go":    true,
		"/realize.go":     true,
		"../test.html":    false,
		"notify.go":       false,
		"realize_test.go": false,
	}
	project := Project{
		Ignore: &Ignore{
			Path: []string{
				"notify.go",
				"*_test.go",
			},
		},
		Watch: &Watch{
			Path: []string{
				"/style.go",
				"./handle.go",
				"../core/*.go",
				"../**/*.go",
				"../**/*.html",
			},
		},
	}
	for p, s := range paths {
		val, _ := project.Validate(p)
		if val != s {
			t.Fatal("Unexpected result", val, "instead", s, "path", p)
		}
	}
	// Test watch extensions and paths
	project = Project{
		Ignore: &Ignore{
			Ext: []string{
				"html",
			},
		},
		Watch: &Watch{
			Ext: []string{
				"go",
			},
			Path: []string{
				"../test/",
			},
		},
	}
	for p := range paths {
		val, _ := project.Validate(p)
		if val {
			t.Fatal("Unexpected result", val, "path", p)
		}
	}
}
