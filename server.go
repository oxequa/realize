package core

import (
	"encoding/json"
	"github.com/tockins/fresh"
	"golang.org/x/net/websocket"
)

type Server struct {
	*Realize `yaml:"-" json:"-"`
	Active   bool        `yaml:"active" json:"active"`
	Port     int         `yaml:"port" json:"port"`
	Host     string      `yaml:"host" json:"host"`
	Server   fresh.Fresh `yaml:"-" json:"-"`
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

func (s *Server) WebSocket(c fresh.Context) (err error) {
	if s.Active {
		ws := c.Request().WS()
		msg, _ := json.Marshal(s.Projects)
		err = websocket.Message.Send(ws, string(msg))
		go func() {
			for {
				select {
				case <-s.Realize.Sync:
					msg, _ := json.Marshal(s.Projects)
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
				err := json.Unmarshal([]byte(text), &s.Projects)
				if err == nil {
					//TODO update config
					break
				}
			}
		}
	}
	return nil
}
