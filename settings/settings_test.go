package settings

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestSettings_Read(t *testing.T) {
	s := Settings{}
	var a interface{}
	s.File = "settings_b"
	if err := s.Read(a); err == nil {
		t.Fatal("Error unexpected", err)
	}

	s.File = "settings_test.yaml"
	dir, err := ioutil.TempDir("", Directory)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ioutil.TempFile(dir, "settings_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s.File = d.Name()
	if err := s.Read(a); err != nil {
		t.Fatal("Error unexpected", err)
	}
}

func TestSettings_Remove(t *testing.T) {
	s := Settings{}
	if err := s.Remove("abcd"); err == nil {
		t.Fatal("Error unexpected, dir dosn't exist", err)
	}

	d, err := ioutil.TempDir("", "settings_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(d); err != nil {
		t.Fatal("Error unexpected, dir exist", err)
	}
}

func TestSettings_Record(t *testing.T) {
	s := Settings{}
	s.File = "settings_test.yaml"
	var a interface{}
	if err := s.Record(a); err != nil {
		t.Fatal(err)
	}
	s.Remove(filepath.Join(Directory, s.File))
}
