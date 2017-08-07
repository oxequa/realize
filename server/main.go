package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/tockins/realize/settings"
	"github.com/tockins/realize/watcher"
	"golang.org/x/net/websocket"
	"gopkg.in/urfave/cli.v2"
)

// Server settings
type Server struct {
	*settings.Settings `yaml:"-"`
	*watcher.Blueprint `yaml:"-"`
	Sync               chan string `yaml:"-"`
}

// Render return a web pages defined in bindata
func render(c echo.Context, path string, mime int) error {
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
	}
	rs.WriteHeader(http.StatusOK)
	rs.Write(data)
	return nil
}

// Start the web server
func (s *Server) Start(p *cli.Context) (err error) {
	if s.Server.Status || p.Bool("server") {
		e := echo.New()
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: 2,
		}))
		e.Use(middleware.Recover())

		// web panel
		e.GET("/", func(c echo.Context) error {
			return render(c, "assets/index.html", 1)
		})
		e.GET("/assets/js/all.min.js", func(c echo.Context) error {
			return render(c, "assets/assets/js/all.min.js", 2)
		})
		e.GET("/assets/css/app.css", func(c echo.Context) error {
			return render(c, "assets/assets/css/app.css", 3)
		})
		e.GET("/app/components/settings/index.html", func(c echo.Context) error {
			return render(c, "assets/app/components/settings/index.html", 1)
		})
		e.GET("/app/components/project/index.html", func(c echo.Context) error {
			return render(c, "assets/app/components/project/index.html", 1)
		})
		e.GET("/app/components/index.html", func(c echo.Context) error {
			return render(c, "assets/app/components/index.html", 1)
		})
		e.GET("/assets/img/svg/ic_settings_black_24px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_settings_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_fullscreen_black_24px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_fullscreen_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_add_black_24px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_add_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_keyboard_backspace_black_24px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_keyboard_backspace_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_error_black_48px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_error_black_48px.svg", 4)
		})
		e.GET("/assets/img/svg/ic_remove_black_24px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_remove_black_24px.svg", 4)
		})
		e.GET("/assets/img/svg/logo.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/logo.svg", 4)
		})
		e.GET("/assets/img/favicon-32x32.png", func(c echo.Context) error {
			return render(c, "assets/assets/img/favicon-32x32.png", 4)
		})
		e.GET("/assets/img/svg/ic_swap_vertical_circle_black_48px.svg", func(c echo.Context) error {
			return render(c, "assets/assets/img/svg/ic_swap_vertical_circle_black_48px.svg", 4)
		})

		//websocket
		e.GET("/ws", s.projects)

		go e.Start(string(s.Settings.Server.Host) + ":" + strconv.Itoa(s.Settings.Server.Port))
		if s.Open || p.Bool("open") {
			_, err = Open("http://" + string(s.Settings.Server.Host) + ":" + strconv.Itoa(s.Settings.Server.Port))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) projects(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		msg, _ := json.Marshal(s.Blueprint.Projects)
		err := websocket.Message.Send(ws, string(msg))
		go func() {
			for {
				select {
				case <-s.Sync:
					msg, _ := json.Marshal(s.Blueprint.Projects)
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
			if err != nil {
				//log.Println(err)
				break
			} else {
				err := json.Unmarshal([]byte(text), &s.Blueprint.Projects)
				if err != nil {
					break
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
