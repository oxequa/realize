package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettings_Flimit(t *testing.T) {
	s := Settings{}
	s.FileLimit = 100
	if err := s.flimit(); err != nil {
		t.Fatal("Unable to increase limit", err)
	}
}

func TestSettings_Stream(t *testing.T) {
	s := Settings{}
	filename := random(4)
	if _, err := s.stream(filename); err == nil {
		t.Fatal("Error expected, none found", filename, err)
	}

	filename = "settings.go"
	if _, err := s.stream(filename); err != nil {
		t.Fatal("Error unexpected", filename, err)
	}
}

func TestSettings_Write(t *testing.T) {
	s := Settings{}
	data := "abcdefgh"
	d, err := ioutil.TempFile("", "io_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.write(d.Name(), []byte(data)); err != nil {
		t.Fatal(err)
	}
}

func TestSettings_Create(t *testing.T) {
	s := Settings{}
	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	f := s.create(p, "io_test")
	os.Remove(f.Name())
}

func TestSettings_Read(t *testing.T) {
	s := Settings{}
	var a interface{}
	s.file = "settings_b"
	if err := s.read(a); err == nil {
		t.Fatal("Error unexpected", err)
	}

	s.file = "settings_test.yaml"
	dir, err := ioutil.TempDir("", directory)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ioutil.TempFile(dir, "settings_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s.file = d.Name()
	if err := s.read(a); err != nil {
		t.Fatal("Error unexpected", err)
	}
}

func TestSettings_Del(t *testing.T) {
	s := Settings{}
	if err := s.del("abcd"); err == nil {
		t.Fatal("Error unexpected, dir dosn't exist", err)
	}

	d, err := ioutil.TempDir("", "settings_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.del(d); err != nil {
		t.Fatal("Error unexpected, dir exist", err)
	}
}

func TestSettings_Record(t *testing.T) {
	s := Settings{}
	s.file = "settings_test.yaml"
	var a interface{}
	if err := s.record(a); err != nil {
		t.Fatal(err)
	}
	s.del(filepath.Join(directory, s.file))
}

func TestSettings_Wdir(t *testing.T) {
	s := Settings{}
	expected, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	result := s.wdir()
	if result != filepath.Base(expected) {
		t.Error("Expected", filepath.Base(expected), "instead", result)
	}
}

func TestSettings_Validate(t *testing.T) {
	s := Settings{}
	input := errors.New("")
	input = nil
	if err := s.validate(input); err != nil {
		t.Error("Expected", input, "instead", err)
	}
}

func TestSettings_Name(t *testing.T) {
	s := Settings{}
	name := random(8)
	path := random(5)
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	result := s.name(name, path)
	if result != dir && result != filepath.Base(path) {
		t.Fatal("Expected", dir, "or", filepath.Base(path), "instead", result)
	}

}

func TestSettings_Path(t *testing.T) {
	s := Settings{}
	path := random(5)
	expected := strings.Replace(filepath.Clean(path), "\\", "/", -1)
	result := s.path(path)
	if result != expected {
		t.Fatal("Expected", expected, "instead", result)
	}

}
