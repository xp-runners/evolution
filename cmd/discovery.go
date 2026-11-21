package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"
)

// splitArgs groups quoted arguments without invoking a shell. Unquoted
// backslashes remain literal, notably in Windows paths and PHP namespaces.
func splitArgs(input string) ([]string, error) {
	var tokens []string
	var word strings.Builder
	var quote rune
	active := false
	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if quote != 0 {
			if r == quote {
				quote = 0
			} else if r == '\\' && i+1 < len(runes) && runes[i+1] == quote {
				i++
				word.WriteRune(runes[i])
			} else {
				word.WriteRune(r)
			}
		} else if unicode.IsSpace(r) {
			if active {
				tokens = append(tokens, word.String())
				word.Reset()
				active = false
			}
		} else if (r == '\'' || r == '"') && !active {
			quote, active = r, true
		} else if r == '"' {
			quote = r
		} else {
			word.WriteRune(r)
			active = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if active {
		tokens = append(tokens, word.String())
	}
	return tokens, nil
}

func systemTimezone() string {
	if zone := strings.TrimPrefix(os.Getenv("TZ"), ":"); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil && zone != "Local" {
			return zone
		}
	}
	if runtime.GOOS != "windows" {
		if target, err := filepath.EvalSymlinks("/etc/localtime"); err == nil {
			if _, zone, ok := strings.Cut(target, "/zoneinfo/"); ok {
				return zone
			}
		}
		if data, err := os.ReadFile("/etc/timezone"); err == nil {
			zone := strings.TrimSpace(string(data))
			if _, err := time.LoadLocation(zone); err == nil && zone != "" && zone != "Local" {
				return zone
			}
		}
	}
	return "UTC"
}

func (c *commandline) list(out io.Writer) (int, error) {
	fmt.Fprintf(out, "\033[33m@%s\033[0m\n", os.Args[0])
	fmt.Fprintln(out, "XP Subcommands")
	fmt.Fprintln(out, "════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(out)

	// Builtin
	fmt.Fprintf(out, "> Builtin @ %s\n\n", c.Version)
	builtins := []string{"list"}
	for name := range commands {
		builtins = append(builtins, name)
	}
	sort.Strings(builtins)
	for _, name := range builtins {
		fmt.Fprintf(out, "  $ xp %s\n", name)
	}

	// Scripts
	if c.project != nil && len(c.project.Scripts) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "> Defined via scripts in @ composer.json")
		var names []string
		for name := range c.project.Scripts {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(out, "  $ xp %s\n", name)
		}
	}

	// Installed
	listDir := func(kind, dir string) {
		files, err := os.ReadDir(filepath.Join(dir, "bin"))
		if err != nil {
			return
		}
		found := false
		for _, file := range files {
			if file.IsDir() || !strings.HasPrefix(file.Name(), "xp.") || strings.HasSuffix(file.Name(), ".bat") {
				continue
			}
			p, err := Plugin("", dir, file.Name())
			if err != nil {
				continue
			}
			f, err := os.Open(filepath.Join(dir, "bin", file.Name()))
			if err != nil {
				continue
			}
			var header [2]byte
			_, err = io.ReadFull(f, header[:])
			f.Close()
			if err != nil || string(header[:]) != "#!" {
				continue
			}
			if !found {
				fmt.Fprintln(out)
				fmt.Fprintf(out, "> %s %s\n\n", kind, dir)
				found = true
			}
			fmt.Fprintf(out, "  $ xp %s (from \033[4;34m%s/%s\033[0m)\n", p.commandName(), p.Name, p.Package)
		}
	}
	for _, m := range c.Modules {
		listDir("Module", m)
	}
	listDir("Local", ".")
	for _, v := range c.vendors() {
		listDir("Installed", v)
	}
	return 0, nil
}
