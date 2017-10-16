package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
	"gopkg.in/urfave/cli.v2"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
)

// Dafault host and port
const (
	host = "localhost"
	port = 5001
)

// Server settings
type Server struct {
	parent *realize
	Status bool   `yaml:"status" json:"status"`
	Open   bool   `yaml:"open" json:"open"`
	Host   string `yaml:"host" json:"host"`
	Port   int    `yaml:"port" json:"port"`
}

// Websocket projects
func (s *Server) projects(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		msg, _ := json.Marshal(s.parent)
		err := websocket.Message.Send(ws, string(msg))
		go func() {
			for {
				select {
				case <-s.parent.sync:
					msg, _ := json.Marshal(s.parent)
					err = websocket.Message.Send(ws, string(msg))
					if err != nil {
						break
					}
				}
			}
		}()
		for {
			// Read
			text := ""
			err := websocket.Message.Receive(ws, &text)
			fmt.Println("receive")
			if err != nil {
				break
			} else {
				err := json.Unmarshal([]byte(text), &s.parent)
				if err == nil {
					s.parent.Settings.record(s.parent)
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

// Start the web server
func (s *Server) start(p *cli.Context) (err error) {
	if p.Bool("server") {
		s.parent.Server.Status = p.Bool("server")
		s.parent.Server.Open = true
	}

	if s.parent.Server.Status {
		e := echo.New()
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 2,
		}))
		e.Use(middleware.Recover())

		// web panel
		e.GET("/", func(c echo.Context) error {
			return s.render(c, "assets/index.html", 1)
		})
		e.GET("/assets/js/all.min.js", func(c echo.Context) error {
			return s.render(c, "assets/assets/js/all.min.js", 2)
		})
		e.GET("/assets/css/app.css", func(c echo.Context) error {
			return s.render(c, "assets/assets/css/app.css", 3)
		})
		e.GET("/app/components/settings/index.html", func(c echo.Context) error {
			return s.render(c, "assets/app/components/settings/index.html", 1)
		})
		e.GET("/app/components/project/index.html", func(c echo.Context) error {
			return s.render(c, "assets/app/components/project/index.html", 1)
		})
		e.GET("/app/components/index.html", func(c echo.Context) error {
			return s.render(c, "assets/app/components/index.html", 1)
		})
		e.GET("/assets/img/svg/ic_settings_black_24px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_settings_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_fullscreen_black_24px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_fullscreen_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_add_black_24px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_add_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_keyboard_backspace_black_24px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_keyboard_backspace_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_error_black_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_error_black_48px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_remove_black_24px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_remove_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/logo.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/logo.svg", 4)
		})
		e.GET("/assets/img/favicon-32x32.png", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/favicon-32x32.png", 5)
		})
		e.GET("/assets/img/svg/ic_swap_vertical_circle_black_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_swap_vertical_circle_black_48px.svg", 4)
		})

		//websocket
		e.GET("/ws", s.projects)
		e.HideBanner = true
		e.Debug = false
		go e.Start(string(s.parent.Server.Host) + ":" + strconv.Itoa(s.parent.Server.Port))
		_, err = s.openURL("http://" + string(s.parent.Server.Host) + ":" + strconv.Itoa(s.parent.Server.Port))
		if err != nil {
			return err
		}
	}
	return nil
}

// OpenURL in a new tab of default browser
func (s *Server) openURL(url string) (io.Writer, error) {
	stderr := bytes.Buffer{}
	cmd := map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
	if s.Open {
		open, err := cmd[runtime.GOOS]
		if !err {
			return nil, fmt.Errorf("operating system %q is not supported", runtime.GOOS)
		}
		cmd := exec.Command(open, url)
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return cmd.Stderr, err
		}
	}
	return nil, nil
}

// Render return a web pages defined in bindata
func (s *Server) render(c echo.Context, path string, mime int) error {
	data, err := Asset(path)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	rs := c.Response()
	// check content type by extensions
	switch mime {
	case 1:
		rs.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		break
	case 2:
		rs.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJavaScriptCharsetUTF8)
		break
	case 3:
		rs.Header().Set(echo.HeaderContentType, "text/css")
		break
	case 4:
		rs.Header().Set(echo.HeaderContentType, "image/svg+xml")
		break
	case 5:
		rs.Header().Set(echo.HeaderContentType, "image/png")
		break
	}
	rs.WriteHeader(http.StatusOK)
	rs.Write(data)
	return nil
}
