package server

//import (
//	"testing"
//	//"fmt"
//	"github.com/tockins/realize/settings"
//	"fmt"
//)
//
//func TestOpen(t *testing.T) {
//	config := settings.Settings{
//		Server: settings.Server{
//			Open: true,
//		},
//	}
//	s := Server{
//		Settings: &config,
//	}
//	url := "open_test"
//	out, err := s.OpenURL(url)
//	if err == nil {
//		t.Fatal("Unexpected, invalid url", url, err)
//	}
//	output := fmt.Sprint(out)
//	if output == "" {
//		t.Fatal("Unexpected, invalid url", url, output)
//	}
//}
