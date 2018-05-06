package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/xp-runners/evolution/composer"
	"github.com/xp-runners/evolution/utf7"
)

var commands = map[string]command{
	"run":     &run{},
	"dump":    &dump{"-d"},
	"write":   &dump{"-w"},
	"eval":    &eval{},
	"version": &version{},
	"help":    &help{},
}

// NewCmd returns an executable command
func NewCmd(c *commandline) (*exec.Cmd, error) {
	main, err := locate(filepath.Join(filepath.Dir(c.Exe), "class-main.php"), "class-main.php")
	if err != nil {
		return nil, err
	}

	command, err := lookupCommand(c)
	if err != nil {
		return nil, err
	}

	// Pass system configuration
	args := []string{
		"-C", "-q",
		"-dmagic_quotes_gpc=0",
		"-dinclude_path=\"" + includePath(*c.Config.Use(), append(c.Modules, command.Modules()...), c.Paths) + "\"",
		"-ddate.timezone=Europe/Berlin", // TBI: Determine this from system TZ
	}
	for name, value := range c.Config.Arguments() {
		args = append(args, "-d"+name+"=\""+value+"\"")
	}
	for _, value := range c.Config.Extensions() {
		args = append(args, "-dextension="+value)
	}

	// Append command arguments
	args = append(args, main)
	for _, arg := range command.Arguments(c.Args) {
		args = append(args, utf7.Encode(arg))
	}

	cmd := exec.Command(*c.Config.Exe(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(
		os.Environ(),
		"XP_EXE="+*c.Config.Exe(),
		"XP_VERSION="+c.Version,
		"XP_MODEL="+c.Model.String(),
		"XP_COMMAND="+c.Command,
	)

	// fmt.Printf("%+v\n", cmd)
	return cmd, nil
}

func locate(files ...string) (string, error) {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
		// Continue
	}

	return "", fmt.Errorf("Failed to locate %v", files)
}

func entry(dir, name string) (string, bool) {
	matches, err := filepath.Glob(filepath.Join(dir, "bin", "xp.*."+name))
	if err != nil || len(matches) == 0 {
		return "", false
	}

	return matches[0], true
}

func lookupCommand(c *commandline) (command, error) {
	if c, ok := commands[c.Command]; ok {
		return c, nil
	}

	for _, dir := range append(c.Modules, composer.Locations()...) {
		if entry, ok := entry(dir, c.Command); ok {
			if c, err := Plugin(c.Command, dir, filepath.Base(entry)); err == nil {
				return c, nil
			}
		}
	}

	return nil, fmt.Errorf("No such command %s", c.Command)
}

func includePath(use string, modules, paths []string) string {
	separator := string(os.PathListSeparator)
	return fmt.Sprintf(
		".%s%s%s%s%s%s.%s%s",
		separator,
		use,
		separator,
		strings.Join(modules, separator),
		separator,
		separator,
		separator,
		strings.Join(paths, separator),
	)
}

func exitStatus(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}

	return 255, err
}
