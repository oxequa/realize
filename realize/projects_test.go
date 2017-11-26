package realize

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func TestProject_After(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	input := "text"
	r.After = func(context Context) {
		log.Println(input)
	}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	r.Projects[0].After()
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Before(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	input := "text"
	r.Before = func(context Context) {
		log.Println(input)
	}
	r.Projects[0].Before()
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}

	r = Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
		Environment: map[string]string{
			input: input,
		},
	})
	r.Projects[0].Before()
	if os.Getenv(input) != input {
		t.Error("Unexpected error expected", input, "instead", os.Getenv(input))
	}
}

func TestProject_Err(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	r := Realize{}
	r.Projects = append(r.Projects, Project{
		parent: &r,
	})
	input := "text"
	r.Err = func(context Context) {
		log.Println(input)
	}
	r.Projects[0].Err(errors.New(input))
	if !strings.Contains(buf.String(), input) {
		t.Error("Unexpected error")
	}
}

func TestProject_Change(t *testing.T) {

}
