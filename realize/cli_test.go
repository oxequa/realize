package realize

import (
	"os"
	"testing"
	"time"
)

type mockRealize struct {
	Settings Settings `yaml:"settings" json:"settings"`
	Server   Server   `yaml:"server" json:"server"`
	Schema   `yaml:",inline"`
	sync     chan string
	exit     chan os.Signal
}

func TestRealize_Stop(t *testing.T) {
	r := Realize{}
	r.exit = make(chan os.Signal, 2)
	r.Stop()
	_, ok := <-r.exit
	if ok != false {
		t.Error("Unexpected error", "channel should be closed")
	}
}

func TestRealize_Start(t *testing.T) {
	r := Realize{}
	go func(){
		time.Sleep(100)
		close(r.exit)
		_, ok := <-r.exit
		if ok != false {
			t.Error("Unexpected error", "channel should be closed")
		}
	}()
	r.Start()
}
