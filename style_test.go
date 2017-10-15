package main

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
	result := red.regular(input)
	c := color.New(color.FgRed).SprintFunc()
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
	result := red.bold(input)
	c := color.New(color.FgRed, color.Bold).SprintFunc()
	expected := fmt.Sprint(c(input))
	if !bytes.Equal([]byte(result), []byte(expected)) {
		t.Error("Expected:", expected, "instead", result)
	}
}
