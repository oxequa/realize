package settings

import (
	"github.com/fatih/color"
)

type Colors struct {
	Red
	Blue
	Yellow
	Magenta
	Green
}
type Red struct{}
type Blue struct{}
type Yellow struct{}
type Magenta struct{}
type Green struct{}

func (c Red) Regular(t ...interface{}) string {
	r := color.New(color.FgRed).SprintFunc()
	return r(t...)
}

func (c Red) Bold(t ...interface{}) string {
	r := color.New(color.FgRed, color.Bold).SprintFunc()
	return r(t...)
}

func (c Blue) Regular(t ...interface{}) string {
	r := color.New(color.FgBlue).SprintFunc()
	return r(t...)
}

func (c Blue) Bold(t ...interface{}) string {
	r := color.New(color.FgBlue, color.Bold).SprintFunc()
	return r(t...)
}

func (c Yellow) Regular(t ...interface{}) string {
	r := color.New(color.FgYellow).SprintFunc()
	return r(t...)
}

func (c Yellow) Bold(t ...interface{}) string {
	r := color.New(color.FgYellow, color.Bold).SprintFunc()
	return r(t...)
}

func (c Magenta) Regular(t ...interface{}) string {
	r := color.New(color.FgMagenta).SprintFunc()
	return r(t...)
}

func (c Magenta) Bold(t ...interface{}) string {
	r := color.New(color.FgMagenta, color.Bold).SprintFunc()
	return r(t...)
}

func (c Green) Regular(t ...interface{}) string {
	r := color.New(color.FgGreen).SprintFunc()
	return r(t...)
}

func (c Green) Bold(t ...interface{}) string {
	r := color.New(color.FgGreen, color.Bold).SprintFunc()
	return r(t...)
}
