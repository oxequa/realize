package main

import (
	"os"
	"strings"
	"testing"
	rc "github.com/tockins/realize/realize"
	"github.com/go-siris/siris/core/errors"
	"bytes"
	"log"
)

var mockResponse interface{}

type mockRealize struct {
	Settings rc.Settings `yaml:"settings" json:"settings"`
	Server   rc.Server   `yaml:"server" json:"server"`
	rc.Schema   `yaml:",inline"`
	sync     chan string
	exit     chan os.Signal
}

func (m *mockRealize) add() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	m.Projects = append(m.Projects, rc.Project{Name:"One"})
	return nil
}

func (m *mockRealize) setup() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) start() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) clean() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) remove() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	m.Projects = []rc.Project{}
	return nil
}

func TestRealize_add(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil{
		t.Fatal("Unexpected error")
	}
	if len(m.Projects) <= 0{
		t.Fatal("Unexpected error")
	}

	m = mockRealize{}
	m.Projects = []rc.Project{{Name:"Default"}}
	mockResponse = nil
	if err := m.add(); err != nil{
		t.Fatal("Unexpected error")
	}
	if len(m.Projects) != 2{
		t.Fatal("Unexpected error")
	}

	m = mockRealize{}
	mockResponse = errors.New("error")
	if err := m.clean(); err == nil{
		t.Fatal("Expected error")
	}
	if len(m.Projects) != 0{
		t.Fatal("Unexpected error")
	}
}

func TestRealize_start(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil{
		t.Fatal("Unexpected error")
	}
}

func TestRealize_setup(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.setup(); err != nil{
		t.Fatal("Unexpected error")
	}
}

func TestRealize_clean(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.clean(); err != nil{
		t.Fatal("Unexpected error")
	}
	mockResponse = errors.New("error")
	if err := m.clean(); err == nil{
		t.Fatal("Expected error")
	}
}

func TestRealize_remove(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.remove(); err != nil{
		t.Fatal("Unexpected error")
	}

	m = mockRealize{}
	mockResponse = nil
	m.Projects = []rc.Project{{Name:"Default"},{Name:"Default"}}
	if err := m.remove(); err != nil{
		t.Fatal("Unexpected error")
	}
	if len(m.Projects) != 0{
		t.Fatal("Unexpected error")
	}

	mockResponse = errors.New("error")
	if err := m.clean(); err == nil{
		t.Fatal("Expected error")
	}
}

func TestRealize_version(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	version()
	if !strings.Contains(buf.String(), rc.RVersion) {
		t.Fatal("Version expted", rc.RVersion)
	}
}