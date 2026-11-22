package main

import (
	"fmt"
	"os"

	"github.com/xp-runners/evolution/cmd"
)

var version = "10.0.0-dev"

func exit(status int, err error) {
	fmt.Fprintf(os.Stderr, "\x1b[33m@%s\x1b[0m\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\x1b[41;37m ERROR \x1b[0;37m %v \x1b[0m", err)
	fmt.Fprintln(os.Stderr)
	os.Exit(status)
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		exit(1, err)
	}

	command, err := cmd.Parse(exe, os.Args[1:], os.Getenv("HOME"), os.Getenv("LOCALAPPDATA"), version)
	if err != nil {
		exit(2, err)
	}

	status, err := command.Run()
	if err != nil {
		exit(status, err)
	}

	os.Exit(status)
}
