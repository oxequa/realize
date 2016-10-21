package config

import (
	"github.com/fatih/color"
)

type Config struct{}

var Red, Blue, Yellow, Magenta = color.New(color.FgRed).SprintFunc(),
	color.New(color.FgBlue).SprintFunc(),
	color.New(color.FgYellow).SprintFunc(),
	color.New(color.FgMagenta).SprintFunc()

var GreenB, RedB, BlueB, YellowB, MagentaB = color.New(color.FgGreen, color.Bold).SprintFunc(),
	color.New(color.FgRed, color.Bold).SprintFunc(),
	color.New(color.FgBlue, color.Bold).SprintFunc(),
	color.New(color.FgYellow, color.Bold).SprintFunc(),
	color.New(color.FgMagenta, color.Bold).SprintFunc()
