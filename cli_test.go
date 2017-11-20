package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"errors"
)

func (m *mockRealize) add() error{
	if mockResponse != nil {
		return mockResponse.(error)
	}
	m.Projects = append(m.Projects, Project{Name:"One"})
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
	m.Projects = []Project{}
	return nil
}

func TestAdd(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil{
		t.Fatal("Unexpected error")
	}
	if len(m.Projects) <= 0{
		t.Fatal("Unexpected error")
	}

	m = mockRealize{}
	m.Projects = []Project{{Name:"Default"}}
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

func TestStart(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil{
		t.Fatal("Unexpected error")
	}
}

func TestSetup(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.setup(); err != nil{
		t.Fatal("Unexpected error")
	}
}

func TestClean(t *testing.T) {
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

func TestRemove(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.remove(); err != nil{
		t.Fatal("Unexpected error")
	}

	m = mockRealize{}
	mockResponse = nil
	m.Projects = []Project{{Name:"Default"},{Name:"Default"}}
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

func TestVersion(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r.version()
	if !strings.Contains(buf.String(), RVersion) {
		t.Fatal("Version expted", RVersion)
	}
}
