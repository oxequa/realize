package main

import (
	"github.com/fatih/color"
)

var (
	output  = color.Output
	red     = colorBase(color.FgRed)
	blue    = colorBase(color.FgBlue)
	green   = colorBase(color.FgGreen)
	yellow  = colorBase(color.FgYellow)
	magenta = colorBase(color.FgMagenta)
)

type colorBase color.Attribute

func (c colorBase) regular(a ...interface{}) string {
	return color.New(color.Attribute(c)).Sprint(a...)
}

func (c colorBase) bold(a ...interface{}) string {
	return color.New(color.Attribute(c), color.Bold).Sprint(a...)
}
