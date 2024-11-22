package logger

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	success = color.New(color.FgGreen, color.Bold).SprintfFunc()
	error   = color.New(color.FgRed, color.Bold).SprintfFunc()
	warning = color.New(color.FgYellow, color.Bold).SprintfFunc()
	info    = color.New(color.FgBlue, color.Bold).SprintfFunc()
	debug   = color.New(color.FgHiBlack).SprintfFunc()
)

func Success(format string, a ...interface{}) {
	fmt.Println(success(format, a...))
}

func Error(format string, a ...interface{}) {
	fmt.Println(error(format, a...))
}

func Warning(format string, a ...interface{}) {
	fmt.Println(warning(format, a...))
}

func Info(format string, a ...interface{}) {
	fmt.Println(info(format, a...))
}

func Debug(format string, a ...interface{}) {
	fmt.Println(debug(format, a...))
}