package realize

import (
	"testing"
	"os"
)

type mockRealize struct {
	Settings Settings `yaml:"settings" json:"settings"`
	Server   Server   `yaml:"server" json:"server"`
	Schema   `yaml:",inline"`
	sync     chan string
	exit     chan os.Signal
}

func TestRealize_Stop(t *testing.T) {
	m := mockRealize{}
	m.exit = make(chan os.Signal, 2)
	close(m.exit)
	_, ok := <-m.exit
	if ok != false {
		t.Error("Unexpected error", "channel should be closed")
	}
}
