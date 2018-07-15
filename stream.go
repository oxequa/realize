package core

import (
	"encoding/json"
	"github.com/tockins/fresh"
	"golang.org/x/net/websocket"
)

type Server struct {
	*Realize
	Port   bool        `yaml:"port,omitempty" json:"port,omitempty"`
	Host   bool        `yaml:"host,omitempty" json:"host,omitempty"`
	Active bool        `yaml:"active,omitempty" json:"active,omitempty"`
	Server fresh.Fresh `yaml:"-" json:"-"`
}

func (s *Server) WebSocket(c fresh.Context) (err error) {
	if s.Active {
		ws := c.Request().WS()
		msg, _ := json.Marshal(s.Schema)
		err = websocket.Message.Send(ws, string(msg))
		go func() {
			for {
				select {
				case <-s.Realize.Sync:
					msg, _ := json.Marshal(s.Schema)
					err = websocket.Message.Send(ws, string(msg))
					if err != nil {
						break
					}
				}
			}
		}()
		for {
			// Read
			var text string
			err = websocket.Message.Receive(ws, &text)
			if err != nil {
				break
			} else {
				err := json.Unmarshal([]byte(text), &s.Schema)
				if err == nil {
					s.Options.Broker.Push(Prefix("Error", Red), err)
					break
				}
			}
		}
	}
	return nil
}

func (s *Server) Start() {
	if s.Active {
		f := fresh.New()
		f.Config().Banner = false
		f.WS("ws", s.WebSocket)
		s.Server = f
		go f.Start()
	}
}
