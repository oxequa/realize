package settings

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSettings_Stream(t *testing.T) {
	s := Settings{}
	filename := Rand(4)
	if _, err := s.Stream(filename); err == nil {
		t.Fatal("Error expected, none found", filename, err)
	}

	filename = "io.go"
	if _, err := s.Stream(filename); err != nil {
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
	if err := s.Write(d.Name(), []byte(data)); err != nil {
		t.Fatal(err)
	}
}

func TestSettings_Create(t *testing.T) {
	s := Settings{}
	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	f := s.Create(p, "io_test")
	os.Remove(f.Name())
}
