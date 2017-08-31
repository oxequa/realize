package settings

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/tockins/realize/style"
	"time"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Wdir return the current working Directory
func (s Settings) Wdir() string {
	dir, err := os.Getwd()
	s.Validate(err)
	return filepath.Base(dir)
}

// Validate checks a fatal error
func (s Settings) Validate(err error) error {
	if err != nil {
		s.Fatal(err, "")
	}
	return nil
}

// Fatal prints a fatal error with its additional messages
func (s Settings) Fatal(err error, msg ...interface{}) {
	if len(msg) > 0 && err != nil {
		log.Fatalln(style.Red.Regular(msg...), err.Error())
	} else if err != nil {
		log.Fatalln(err.Error())
	}
}

// Name return the project name or the path of the working dir
func (s Settings) Name(name string, path string) string {
	if name == "" && path == "" {
		return s.Wdir()
	} else if path != "/" {
		return filepath.Base(path)
	}
	return name
}

// Path cleaner
func (s Settings) Path(path string) string {
	return strings.Replace(filepath.Clean(path), "\\", "/", -1)
}

// Rand is used for generate a random string
func Rand(n int) string {
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
