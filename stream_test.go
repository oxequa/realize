package core

import (
	"testing"
	"time"
)

func TestStream_Start(t *testing.T) {
	stream := Server{Active: true}
	go stream.Start()
	func() {
		time.Sleep(100 * time.Millisecond)
		stream.Server.Stop()
	}()
}
