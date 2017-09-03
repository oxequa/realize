package main

import (
	"flag"
	"fmt"
	"github.com/tockins/realize/settings"
	"github.com/tockins/realize/style"
	"github.com/tockins/realize/watcher"
	"gopkg.in/urfave/cli.v2"
	"testing"
)

func TestPrefix(t *testing.T) {
	input := settings.Rand(10)
	value := fmt.Sprint(style.Yellow.Bold("[")+"REALIZE"+style.Yellow.Bold("]"), input)
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

func TestNoConf(t *testing.T) {
	settings := settings.Settings{Make: true}
	set := flag.NewFlagSet("test", 0)
	set.Bool("no-config", true, "")
	params := cli.NewContext(nil, set, nil)
	noconf(params, &settings)
	if settings.Make == true {
		t.Fatal("Expected", false, "Instead", true)
	}
}

func TestPolling(t *testing.T) {
	settings := settings.Legacy{}
	set := flag.NewFlagSet("test", 0)
	set.Bool("legacy", true, "")
	params := cli.NewContext(nil, set, nil)
	polling(params, &settings)
	if settings.Interval == 0 {
		t.Fatal("Expected interval", settings.Interval, "Instead", 0)
	}
}

func TestInsert(t *testing.T) {
	b := watcher.Blueprint{}
	b.Settings = &settings.Settings{}
	set := flag.NewFlagSet("test", 0)
	set.String("name", settings.Rand(5), "")
	set.String("path", settings.Rand(5), "")
	params := cli.NewContext(nil, set, nil)
	if err := insert(params, &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Projects) == 0 {
		t.Error("Expected one project")
	}
}
