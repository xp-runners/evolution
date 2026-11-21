// Command tput implements the XP Runners subset of tput.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var version = "9.4.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, terminalSize))
}

func run(args []string, in io.Reader, out, stderr io.Writer, size func() (cols, lines int)) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: tput [-V] [-S] capname")
		return 1
	}
	if args[0] == "-V" {
		fmt.Fprintln(out, "xp-runners", version)
		return 0
	}

	// Supported subset of capabilities
	capabilities := map[string]func() string{
		"lines": func() string {
			_, lines := size()
			return strconv.Itoa(lines)
		},
		"cols": func() string {
			cols, _ := size()
			return strconv.Itoa(cols)
		},
		"bel": func() string {
			return "\a"
		},
	}

	// Read capabilities from standard input when -S is given
	var queries []string
	if args[0] == "-S" {
		queries = make([]string, 0, 1)
		reader := bufio.NewReader(in)
		for {
			line, err := reader.ReadString('\n')
			if len(line) != 0 {
				queries = append(queries, strings.TrimSpace(line))
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintf(stderr, "tput: %v\n", err)
				return 1
			}
		}
	} else {
		queries = args[:1]
	}

	for _, query := range queries {
		if capability, exists := capabilities[query]; exists {
			fmt.Fprintln(out, capability())
		} else {
			fmt.Fprintf(stderr, "tput: unknown terminfo capability '%s'\n", query)
			return 1
		}
	}
	return 0
}

func terminalSize() (cols, lines int) {
	// Stdout is often redirected, for example in cols=$(tput cols).
	for _, f := range []*os.File{os.Stdout, os.Stderr, os.Stdin} {
		if cols, lines := windowSize(f); cols > 0 && lines > 0 {
			return cols, lines
		}
	}
	if f, err := os.Open(consoleDevice); err == nil {
		defer f.Close()
		return windowSize(f)
	}
	return 0, 0
}
