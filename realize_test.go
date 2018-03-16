package main

import (
	"bytes"
	"errors"
	"github.com/oxequa/realize/realize"
	"log"
	"strings"
	"testing"
)

var mockResponse interface{}

type mockRealize realize.Realize

func (m *mockRealize) add() error {
	if mockResponse != nil {
		return mockResponse.(error)
	}
	m.Projects = append(m.Projects, realize.Project{Name: "One"})
	return nil
}

func (m *mockRealize) setup() error {
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) start() error {
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) clean() error {
	if mockResponse != nil {
		return mockResponse.(error)
	}
	return nil
}

func (m *mockRealize) remove() error {
	if mockResponse != nil {
		return mockResponse.(error)
	}
	m.Projects = []realize.Project{}
	return nil
}

func TestRealize_add(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil {
		t.Error("Unexpected error")
	}
	if len(m.Projects) <= 0 {
		t.Error("Unexpected error")
	}

	m = mockRealize{}
	m.Projects = []realize.Project{{Name: "Default"}}
	mockResponse = nil
	if err := m.add(); err != nil {
		t.Error("Unexpected error")
	}
	if len(m.Projects) != 2 {
		t.Error("Unexpected error")
	}

	m = mockRealize{}
	mockResponse = errors.New("error")
	if err := m.clean(); err == nil {
		t.Error("Expected error")
	}
	if len(m.Projects) != 0 {
		t.Error("Unexpected error")
	}
}

func TestRealize_start(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.add(); err != nil {
		t.Error("Unexpected error")
	}
}

func TestRealize_setup(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.setup(); err != nil {
		t.Error("Unexpected error")
	}
}

func TestRealize_clean(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.clean(); err != nil {
		t.Error("Unexpected error")
	}
	mockResponse = errors.New("error")
	if err := m.clean(); err == nil {
		t.Error("Expected error")
	}
}

func TestRealize_remove(t *testing.T) {
	m := mockRealize{}
	mockResponse = nil
	if err := m.remove(); err != nil {
		t.Error("Unexpected error")
	}

	m = mockRealize{}
	mockResponse = nil
	m.Projects = []realize.Project{{Name: "Default"}, {Name: "Default"}}
	if err := m.remove(); err != nil {
		t.Error("Unexpected error")
	}
	if len(m.Projects) != 0 {
		t.Error("Unexpected error")
	}

	mockResponse = errors.New("error")
	if err := m.clean(); err == nil {
		t.Error("Expected error")
	}
}

func TestRealize_version(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	version()
	if !strings.Contains(buf.String(), realize.RVersion) {
		t.Error("Version expted", realize.RVersion)
	}
}
