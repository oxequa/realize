package server

import (
	"testing"
	"github.com/tockins/realize/settings"
	"net/http"
)

func TestServer_Start(t *testing.T) {
	s := settings.Settings{
		Server: settings.Server{
			Status: true,
			Open:   false,
			Host: "localhost",
			Port: 5000,
		},
	}
	server := Server{
		Settings:  &s,
	}
	err := server.Start(nil)
	if err != nil{
		t.Fatal(err)
	}
	host := "http://localhost:5000/"
	urls := []string{
		host,
		host+"ws",
		host+"assets/js/all.min.js",
		host+"assets/css/app.css",
		host+"app/components/settings/index.html",
		host+"app/components/project/index.html",
		host+"app/components/project/index.html",
		host+"app/components/index.html",
		host+"assets/img/svg/ic_settings_black_24px.svg",
		host+"assets/img/svg/ic_fullscreen_black_24px.svg",
		host+"assets/img/svg/ic_add_black_24px.svg",
		host+"assets/img/svg/ic_keyboard_backspace_black_24px.svg",
		host+"assets/img/svg/ic_error_black_48px.svg",
		host+"assets/img/svg/ic_remove_black_24px.svg",
		host+"assets/img/svg/logo.svg",
		host+"assets/img/favicon-32x32.png",
		host+"assets/img/svg/ic_swap_vertical_circle_black_48px.svg",
	}
	for _, elm := range urls {
		resp, err := http.Get(elm)
		if err != nil || resp.StatusCode != 200 {
			t.Fatal(err)
		}
	}
}
