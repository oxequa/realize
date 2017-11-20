package realize

import (
	"github.com/fatih/color"
)

var (
	Output  = color.Output
	Red     = colorBase(color.FgHiRed)
	Blue    = colorBase(color.FgHiBlue)
	Green   = colorBase(color.FgHiGreen)
	Yellow  = colorBase(color.FgHiYellow)
	Magenta = colorBase(color.FgHiMagenta)
)

// ColorBase type
type colorBase color.Attribute

// Regular font with a color
func (c colorBase) Regular(a ...interface{}) string {
	return color.New(color.Attribute(c)).Sprint(a...)
}

// Bold font with a color
func (c colorBase) Bold(a ...interface{}) string {
	return color.New(color.Attribute(c), color.Bold).Sprint(a...)
}
