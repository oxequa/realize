package style

import (
	"fmt"
	"bytes"
	"testing"
)

func TestColorBase_Regular(t *testing.T) {
	c := new(colorBase)
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	for i, s := range strs {
		input[i] = s
	}
	result := c.Regular(input)
	expected := fmt.Sprint(input)
	if !bytes.Equal([]byte(result), []byte(expected)){
		t.Error("Expected:", expected, "instead", result)
	}
}

func TestColorBase_Bold(t *testing.T) {
	c := new(colorBase)
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	for i, s := range strs {
		input[i] = s
	}
	result := c.Bold(input)
	expected := fmt.Sprint(input)
	if !bytes.Equal([]byte(result), []byte(expected)){
		t.Error("Expected:", expected, "instead", result)
	}
}
