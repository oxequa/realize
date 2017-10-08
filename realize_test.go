package main

//import (
//	"flag"
//	"fmt"
//	"github.com/tockins/realize/settings"
//	"github.com/tockins/realize/style"
//	"github.com/tockins/realize/watcher"
//	"gopkg.in/urfave/cli.v2"
//	"testing"
//)
//
//func TestPrefix(t *testing.T) {
//	input := random(10)
//	value := fmt.Sprint(yellow.bold("[")+"REALIZE"+yellow.bold("]"), input)
//	result := prefix(input)
//	if result == "" {
//		t.Fatal("Expected a string")
//	}
//	if result != value {
//		t.Fatal("Expected", value, "Instead", result)
//	}
//}
//
//func TestBefore(t *testing.T) {
//	context := cli.Context{}
//	if err := before(&context); err != nil {
//		t.Fatal(err)
//	}
//}
//
//func TestInsert(t *testing.T) {
//	b := Blueprint{}
//	b.Settings = &Settings{}
//	set := flag.NewFlagSet("test", 0)
//	set.String("name", random(5), "")
//	set.String("path", random(5), "")
//	params := cli.NewContext(nil, set, nil)
//	if err := insert(params, &b); err != nil {
//		t.Fatal(err)
//	}
//	if len(b.Projects) == 0 {
//		t.Error("Expected one project")
//	}
//}
