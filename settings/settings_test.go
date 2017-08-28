package settings

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestSettings_Read(t *testing.T) {
	s := Settings{}
	var a interface{}
	s.Resources.Config = "settings_b"
	if err := s.Read(a); err == nil {
		t.Fatal("Error unexpected", err)
	}

	s.Resources.Config = "settings_test.yaml"
	d, err := ioutil.TempFile("", "settings_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s.Resources.Config = d.Name()
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
	s.Resources.Config = "settings_test.yaml"
	var a interface{}
	if err := s.Record(a); err != nil {
		t.Fatal(err)
	}
	s.Remove(filepath.Join(directory, s.Resources.Config))
}
