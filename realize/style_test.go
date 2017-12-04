package realize

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"testing"
)

func TestStyle_Regular(t *testing.T) {
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	for i, s := range strs {
		input[i] = s
	}
	result := Red.Regular(input)
	c := color.New(color.FgHiRed).SprintFunc()
	expected := fmt.Sprint(c(input))
	if !bytes.Equal([]byte(result), []byte(expected)) {
		t.Error("Expected:", expected, "instead", result)
	}
}

func TestStyle_Bold(t *testing.T) {
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	for i, s := range strs {
		input[i] = s
	}
	result := Red.Bold(input)
	c := color.New(color.FgHiRed, color.Bold).SprintFunc()
	expected := fmt.Sprint(c(input))
	if !bytes.Equal([]byte(result), []byte(expected)) {
		t.Error("Expected:", expected, "instead", result)
	}
}
