package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	c "github.com/tockins/realize/cli"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

var Bp *c.Blueprint

// Server struct contains server informations
type Server struct {
}

func render(c echo.Context, path string) error {
	data, err := Asset(path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	rs := c.Response()
	rs.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	rs.WriteHeader(http.StatusOK)
	rs.Write(data)
	return nil
}

func (s *Server) Start() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.GET("/", func(c echo.Context) error {
		return render(c, "server/assets/index.html")
	})

	e.GET("/projects", standard.WrapHandler(projects()))
	go e.Run(standard.New(":5000"))
}

// The WebSocket for projects list
func projects() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		for {
			err := websocket.Message.Send(ws, "Hello")
			if err != nil {
				log.Fatal(err)
			}
			msg := ""
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				log.Fatal(err)
			}
		}
	})
}
