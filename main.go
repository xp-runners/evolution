package main

import (
	"fmt"
	"os"

	"github.com/xp-runners/evolution/cmd"
)

var version = "9.4.0-dev"

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	command, err := cmd.Parse(exe, os.Args[1:], os.Getenv("HOME"), os.Getenv("LOCALAPPDATA"), version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	status, err := command.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(status)
	}

	os.Exit(status)
}
