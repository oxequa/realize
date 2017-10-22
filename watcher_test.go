package main

import (
	"testing"

	"os"
)

type fileWatcherMock struct {
	FileWatcher
}

func (f *fileWatcherMock) Walk(path string, _ bool) string {
	return path
}

type fileInfoMock struct {
	os.FileInfo
	FileIsDir bool
}

func (m *fileInfoMock) IsDir() bool { return m.FileIsDir }

func TestWalk(t *testing.T) {
	p := Project{
		Name: "Test Project",
		Watcher: Watch{
			Paths:  []string{"/"},
			Ignore: []string{"vendor"},
			Exts:   []string{"go"},
		},
		Path:    "/go/project",
		watcher: &fileWatcherMock{},
		init:    true,
	}

	files := []struct {
		Path  string
		IsDir bool
	}{
		// valid files
		{"/go/project", true},
		{"/go/project/main.go", false},
		{"/go/project/main_test.go", false},
		// invalid relative path
		{"./relative/path", true},
		{"./relative/path/file.go", false},
		// invalid extension
		{"/go/project/settings.yaml", false},
		// invalid vendor files
		{"/go/project/vendor/foo", true},
		{"/go/project/vendor/foo/main.go", false},
	}

	for _, file := range files {
		fileInfo := fileInfoMock{
			FileIsDir: file.IsDir,
		}
		err := p.walk(file.Path, &fileInfo, nil)
		if err != nil {
			t.Errorf("Error not expected: %s", err)
		}
	}

	if p.files != 2 {
		t.Errorf("Exepeted %d files, but was %d", 2, p.files)
	}

	if p.folders != 1 {
		t.Errorf("Exepeted %d folders, but was %d", 2, p.folders)
	}
}
