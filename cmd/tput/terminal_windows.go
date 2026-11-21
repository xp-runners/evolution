package main

import (
	"os"

	"golang.org/x/sys/windows"
)

const consoleDevice = "CONOUT$"

func windowSize(f *os.File) (cols, lines int) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(f.Fd()), &info); err != nil {
		return 0, 0
	}
	return int(info.Window.Right) - int(info.Window.Left) + 1,
		int(info.Window.Bottom) - int(info.Window.Top) + 1
}
