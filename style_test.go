package main

//
//import (
//	"bytes"
//	"fmt"
//	"github.com/fatih/color"
//	"testing"
//)
//
//func TestStyle_Regular(t *testing.T) {
//	strs := []string{"a", "b", "c"}
//	input := make([]interface{}, len(strs))
//	for i, s := range strs {
//		input[i] = s
//	}
//	result := style.Regular(input)
//	expected := fmt.Sprint(input)
//	if !bytes.Equal([]byte(result), []byte(expected)) {
//		t.Error("Expected:", expected, "instead", result)
//	}
//}
//
//func TestStyle_Bold(t *testing.T) {
//	strs := []string{"a", "b", "c"}
//	input := make([]interface{}, len(strs))
//	for i, s := range strs {
//		input[i] = s
//	}
//	result := style.Bold(input)
//	expected := fmt.Sprint(input)
//	if !bytes.Equal([]byte(result), []byte(expected)) {
//		t.Error("Expected:", expected, "instead", result)
//	}
//}
