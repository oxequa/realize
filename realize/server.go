//go:generate go-bindata -pkg=realize -o=bindata.go assets/...

package realize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"github.com/go-siris/siris/core/errors"
)

// Dafault host and port
const (
	Host = "localhost"
	Port = 5002
)

// Server settings
type Server struct {
	Parent *Realize `yaml:"-" json:"-"`
	Status bool     `yaml:"status" json:"status"`
	Open   bool     `yaml:"open" json:"open"`
	Port   int      `yaml:"port" json:"port"`
	Host   string   `yaml:"host" json:"host"`
}

// Websocket projects
func (s *Server) projects(c echo.Context) (err error) {
	websocket.Handler(func(ws *websocket.Conn) {
		msg, _ := json.Marshal(s.Parent)
		err = websocket.Message.Send(ws, string(msg))
		go func() {
			for {
				select {
				case <-s.Parent.Sync:
					msg, _ := json.Marshal(s.Parent)
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
			err = websocket.Message.Receive(ws, &text)
			if err != nil {
				break
			} else {
				err := json.Unmarshal([]byte(text), &s.Parent)
				if err == nil {
					s.Parent.Settings.Write(s.Parent)
					break
				}
			}
		}
		ws.Close()
	}).ServeHTTP(c.Response(), c.Request())
	return nil
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

func (s *Server) Set(status bool, open bool, port int, host string) {
	s.Open = open
	s.Port = port
	s.Host = host
	s.Status = status
}

// Start the web server
func (s *Server) Start() (err error) {
	if s.Status {
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
		e.GET("/assets/img/logo.png", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/logo.png", 5)
		})
		e.GET("/assets/img/svg/github-logo.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/github-logo.svg", 4)
		})
		e.GET("/assets/img/svg/ic_arrow_back_black_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_arrow_back_black_48px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_clear_white_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_clear_white_48px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_menu_white_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_menu_white_48px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_settings_black_48px.svg", func(c echo.Context) error {
			return s.render(c, "assets/assets/img/svg/ic_settings_black_48px.svg", 4)
		})

		//websocket
		e.GET("/ws", s.projects)
		e.HideBanner = true
		e.Debug = false
		go func() {
			log.Println(s.Parent.Prefix("Started on " + string(s.Host) + ":" + strconv.Itoa(s.Port)))
			e.Start(string(s.Host) + ":" + strconv.Itoa(s.Port))
		}()
	}
	return nil
}

// OpenURL in a new tab of default browser
func (s *Server) OpenURL() error {
	url := "http://" + string(s.Parent.Server.Host) + ":" + strconv.Itoa(s.Parent.Server.Port)
	stderr := bytes.Buffer{}
	cmd := map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}
	if s.Open {
		open, err := cmd[runtime.GOOS]
		if !err {
			return fmt.Errorf("operating system %q is not supported", runtime.GOOS)
		}
		cmd := exec.Command(open, url)
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.New(stderr.String())
		}
	}
	return nil
}
