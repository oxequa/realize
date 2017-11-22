package realize

import (
	"testing"
	"os"
)

func TestProject_Setup(t *testing.T) {
	input := "Rtest"
	p := Project{
		Path: "/test/prova/"+input,
		Environment: map[string]string{
			input: input,
		},
	}
	p.Setup()
	if p.Name != input{
		t.Error("Unexpected error", p.Name,"instead",input)
	}
	if os.Getenv(input) != input{
		t.Error("Unexpected error", os.Getenv(input),"instead",input)
	}
}