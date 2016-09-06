package server

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func render(c echo.Context, path string) error{
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

func init() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.GET("/", func(c echo.Context) error {
		return render(c, "server/assets/index.html")
	})
	go e.Run(standard.New(":5000"))
}
