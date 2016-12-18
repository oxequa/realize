package settings

import (
	"github.com/fatih/color"
)

// Colors allowed
type Colors struct {
	Red
	Blue
	Yellow
	Magenta
	Green
}

// Red color
type Red struct{}

// Blue color
type Blue struct{}

// Yellow color
type Yellow struct{}

// Magenta color
type Magenta struct{}

// Green color
type Green struct{}

// Regular font in red
func (c Red) Regular(t ...interface{}) string {
	r := color.New(color.FgRed).SprintFunc()
	return r(t...)
}

// Bold font in red
func (c Red) Bold(t ...interface{}) string {
	r := color.New(color.FgRed, color.Bold).SprintFunc()
	return r(t...)
}

// Regular font in blue
func (c Blue) Regular(t ...interface{}) string {
	r := color.New(color.FgBlue).SprintFunc()
	return r(t...)
}

// Bold font in blue
func (c Blue) Bold(t ...interface{}) string {
	r := color.New(color.FgBlue, color.Bold).SprintFunc()
	return r(t...)
}

// Regular font in yellow
func (c Yellow) Regular(t ...interface{}) string {
	r := color.New(color.FgYellow).SprintFunc()
	return r(t...)
}

// Bold font in red
func (c Yellow) Bold(t ...interface{}) string {
	r := color.New(color.FgYellow, color.Bold).SprintFunc()
	return r(t...)
}

// Regular font in magenta
func (c Magenta) Regular(t ...interface{}) string {
	r := color.New(color.FgMagenta).SprintFunc()
	return r(t...)
}

// Bold font in magenta
func (c Magenta) Bold(t ...interface{}) string {
	r := color.New(color.FgMagenta, color.Bold).SprintFunc()
	return r(t...)
}

// Regular font in green
func (c Green) Regular(t ...interface{}) string {
	r := color.New(color.FgGreen).SprintFunc()
	return r(t...)
}

// Bold font in red
func (c Green) Bold(t ...interface{}) string {
	r := color.New(color.FgGreen, color.Bold).SprintFunc()
	return r(t...)
}
