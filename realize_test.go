package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v2"
	"reflect"
	"testing"
)

func TestPrefix(t *testing.T) {
	input := random(10)
	value := fmt.Sprint(yellow.bold("[")+"REALIZE"+yellow.bold("]"), input)
	result := prefix(input)
	if result == "" {
		t.Fatal("Expected a string")
	}
	if result != value {
		t.Fatal("Expected", value, "Instead", result)
	}
}

func TestBefore(t *testing.T) {
	context := cli.Context{}
	if err := before(&context); err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	r := new()
	t.Log(reflect.TypeOf(r).String())
	if reflect.TypeOf(r).String() != "main.realize" {
		t.Error("Expected a realize struct")
	}
}
