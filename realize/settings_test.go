package realize

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

// Rand is used for generate a random string
func random(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func TestSettings_Stream(t *testing.T) {
	s := Settings{}
	filename := random(4)
	if _, err := s.Stream(filename); err == nil {
		t.Fatal("Error expected, none found", filename, err)
	}

	filename = "settings.go"
	if _, err := s.Stream(filename); err != nil {
		t.Fatal("Error unexpected", filename, err)
	}
}

func TestSettings_Remove(t *testing.T) {
	s := Settings{}
	if err := s.Remove("abcd"); err == nil {
		t.Fatal("Error unexpected, dir dosn't exist", err)
	}

	d, err := ioutil.TempDir("", "settings_test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(d); err != nil {
		t.Fatal("Error unexpected, dir exist", err)
	}
}

func TestSettings_Write(t *testing.T) {
	s := Settings{}
	data := "abcdefgh"
	d, err := ioutil.TempFile("", "io_test")
	if err != nil {
		t.Fatal(err)
	}
	RFile = d.Name()
	if err := s.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
}

func TestSettings_Read(t *testing.T) {
	s := Settings{}
	var a interface{}
	RFile = "settings_b"
	if err := s.Read(a); err == nil {
		t.Fatal("Error unexpected", err)
	}
	RFile = "settings_test.yaml"
	d, err := ioutil.TempFile("", "settings_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	RFile = d.Name()
	if err := s.Read(a); err != nil {
		t.Fatal("Error unexpected", err)
	}
}

func TestSettings_Fatal(t *testing.T) {
	s := Settings{}
	s.Fatal(nil, "test")
}

func TestSettings_Create(t *testing.T) {
	s := Settings{}
	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	f := s.Create(p, "io_test")
	os.Remove(f.Name())
}
