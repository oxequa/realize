package server

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	c "github.com/tockins/realize/cli"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

// Server struct contains server informations
type Server struct {
	Blueprint *c.Blueprint
	Files     map[string]string
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

// Server starting
func (s *Server) Start() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, s.Blueprint)
		//return render(c, "server/assets/index.html")
	})

	e.GET("/projects", standard.WrapHandler(projects()))
	go e.Run(standard.New(":5000"))
}

// The WebSocket for projects list
func projects() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		for {
			message, _ := json.Marshal("")
			err := websocket.Message.Send(ws, string(message))
			fmt.Println("")
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
