package core

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestActivityWalk(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var watcher FileWatcher
	watcher, err = NewFileWatcher(Legacy{Force: false})
	if err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs(dir)
	path = filepath.Join(dir, string(os.PathSeparator))
	model := Activity{Ignore: &Ignore{Hidden: true}}
	if err := model.Walk(path, watcher); err != nil {
		t.Fatal(err)
	}
	var countFiles int
	var countFolders int
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if !file.IsDir() && !Hidden(file.Name()) {
			countFiles++
		} else if file.IsDir() {
			countFolders++
		}
	}
	if len(model.Files) != countFiles {
		t.Fatal("Wrong files count")
	}
	if len(model.Folders) != countFolders {
		t.Fatal("Wrong folders count")
	}
}

func TestActivityExec(t *testing.T) {
	var buf bytes.Buffer
	var wg sync.WaitGroup
	log.SetOutput(&buf)
	wg.Add(1)
	activity := Activity{}
	reload := make(chan bool)
	command := Command{Cmd: "sleep 1"}
	activity.Exec(command, &wg, reload)
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
	activity := Activity{
		Realize: &realize,
		Watch: &Watch{
			Path: []string{
				"../**/*.go",
			},
		},
	}
	sequence := Series{
		Tasks: toInterface([]Command{
			{
				Cmd: "echo test",
			},
			{
				Cmd: "sleep 1",
			},
		}),
	}
	activity.Tasks = make([]interface{}, 0)
	activity.Tasks = append(activity.Tasks, sequence)
	// stop scan after 1.5 sec
	go func() {
		time.Sleep(1500 * time.Millisecond)
		realize.Exit <- true
	}()
	activity.Scan(&wg)
	if buf.Len() == 0 {
		t.Fatal("Unexpected error")
	}

}

func TestActivityReload(t *testing.T) /**/ {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	activity := Activity{}
	reload := make(chan bool)
	tasks := make([]interface{}, 0)
	parallel := Parallel{
		Tasks: toInterface([]Command{
			{
				Cmd: "echo clean super command root test",
			},
			{
				Cmd: "go fmt",
			},
		}),
	}
	sequence := Series{
		Tasks: toInterface([]Command{
			{
				Cmd: "go install",
			},
			{
				Cmd: "go build",
			},
		}),
	}
	tasks = append(tasks, parallel)
	tasks = append(tasks, sequence)
	activity.Run(reload, tasks...)
	if buf.Len() == 0 {
		t.Fatal("Unexpected error")
	}
}

func TestActivityValidate(t *testing.T) {
	// Test paths
	paths := map[string]bool{
		"/style.go":       true,
		"./handle.go":     true,
		"/options.go":     true,
		"/realize.go":     true,
		"../test.html":    false,
		"notify.go":       false,
		"realize_test.go": false,
	}
	activity := Activity{
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
		val, _ := activity.Validate(p)
		if val != s {
			t.Fatal("Unexpected result", val, "instead", s, "path", p)
		}
	}
	// Test watch extensions and paths
	activity = Activity{
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
		val, _ := activity.Validate(p)
		if val {
			t.Fatal("Unexpected result", val, "path", p)
		}
	}
}
