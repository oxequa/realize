package main

import (
	"testing"
	"time"
)

func TestProject_GoCompile(t *testing.T) {
	p := Project{}
	stop := make(chan bool)
	response := make(chan string)
	result, err := p.goCompile(stop, []string{"echo"}, []string{"test"})
	if err != nil {
		t.Error("Unexpected", err)
	}
	go func() {
		result, _ = p.goCompile(stop, []string{"sleep"}, []string{"20s"})
		response <- result
	}()
	close(stop)
	select {
	case v := <-response:
		if v != msgStop {
			t.Error("Unexpected result", response)
		}
	case <-time.After(2 * time.Second):
		t.Error("Channel doesn't works")
	}
}
