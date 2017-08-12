package settings

import (
	"errors"
	"github.com/labstack/gommon/random"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettings_Wdir(t *testing.T) {
	s := Settings{}
	expected, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	result := s.Wdir()
	if result != filepath.Base(expected) {
		t.Error("Expected", filepath.Base(expected), "instead", result)
	}
}

func TestSettings_Validate(t *testing.T) {
	s := Settings{}
	input := errors.New("")
	input = nil
	if err := s.Validate(input); err != nil {
		t.Error("Expected", input, "instead", err)
	}
}

func TestSettings_Name(t *testing.T) {
	s := Settings{}
	name := random.String(8)
	path := random.String(5)
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	result := s.Name(name, path)
	if result != dir && result != filepath.Base(path) {
		t.Fatal("Expected", dir, "or", filepath.Base(path), "instead", result)
	}

}

func TestSettings_Path(t *testing.T) {
	s := Settings{}
	path := random.String(5)
	expected := strings.Replace(filepath.Clean(path), "\\", "/", -1)
	result := s.Path(path)
	if result != expected {
		t.Fatal("Expected", expected, "instead", result)
	}

}
