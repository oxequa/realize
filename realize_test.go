package main

import "os"

var mockResponse interface{}

type mockRealize struct {
	Settings Settings `yaml:"settings" json:"settings"`
	Server   Server   `yaml:"server" json:"server"`
	Schema   `yaml:",inline"`
	sync     chan string
	exit     chan os.Signal
}
