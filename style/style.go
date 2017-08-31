package style

import (
	"github.com/fatih/color"
)

type colorBase color.Attribute

func (s colorBase) Regular(a ...interface{}) string {
	return color.New(color.Attribute(s)).Sprint(a...)
}

func (s colorBase) Bold(a ...interface{}) string {
	return color.New(color.Attribute(s), color.Bold).Sprint(a...)
}

// allowed colors
var (
	Red     = colorBase(color.FgRed)
	Blue    = colorBase(color.FgBlue)
	Yellow  = colorBase(color.FgYellow)
	Magenta = colorBase(color.FgMagenta)
	Green   = colorBase(color.FgGreen)
)

// Output defines the standard output of the print functions. By default, os.Stdout is used.
// When invoked on Windows machines, this automatically handles escape sequences.
var Output = color.Output
