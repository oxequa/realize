package main

import (
	"github.com/fatih/color"
)

var (
	output  = color.Output
	red     = colorBase(color.FgHiRed)
	blue    = colorBase(color.FgHiBlue)
	green   = colorBase(color.FgHiGreen)
	yellow  = colorBase(color.FgHiYellow)
	magenta = colorBase(color.FgHiMagenta)
)

type colorBase color.Attribute

func (c colorBase) regular(a ...interface{}) string {
	return color.New(color.Attribute(c)).Sprint(a...)
}

func (c colorBase) bold(a ...interface{}) string {
	return color.New(color.Attribute(c), color.Bold).Sprint(a...)
}
