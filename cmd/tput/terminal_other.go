//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !windows

package main

import "os"

const consoleDevice = ""

func windowSize(*os.File) (cols, lines int) { return 0, 0 }
