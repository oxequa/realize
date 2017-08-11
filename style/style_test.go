package style

import (
	"testing"
	"fmt"
)

func TestColorBase_Regular(t *testing.T) {
	c := new(colorBase)
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	result := c.Bold(input)
	expected := fmt.Sprint(input)
	for i, s := range strs {
		input[i] = s
	}
	if result != expected{
		t.Error("Expected:", expected, "instead", result)
	}
}

func TestColorBase_Bold(t *testing.T) {
	c := new(colorBase)
	strs := []string{"a", "b", "c"}
	input := make([]interface{}, len(strs))
	result := c.Bold(input)
	expected := fmt.Sprint(input)
	for i, s := range strs {
		input[i] = s
	}
	if result != expected{
		t.Error("Expected:", expected, "instead", result)
	}
}