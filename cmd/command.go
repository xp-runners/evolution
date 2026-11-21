package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
	"ar":      &entryCommand{"xp.xar.Runner"},
}

type entryCommand struct{ class string }

func (e entryCommand) Arguments(args []string) []string { return append([]string{e.class}, args...) }
func (e entryCommand) Modules() []string                { return nil }

// NewCmd returns an executable command
func NewCmd(c *commandline) (*exec.Cmd, error) {
	main, err := locate(filepath.Join(filepath.Dir(c.Exe), "class-main.php"), "class-main.php")
	if err != nil {
		return nil, err
	}

	resolved, err := lookupCommand(c)
	if err != nil {
		return nil, err
	}

	config, err := c.settings()
	if err != nil {
		return nil, err
	}
	runtime, err := splitArgs(config.executable)
	if err != nil || len(runtime) == 0 || runtime[0] == "" {
		return nil, fmt.Errorf("invalid PHP runtime %q", config.executable)
	}
	if config.use == nil {
		for _, vendor := range c.vendors() {
			path := filepath.Join(vendor, "xp-framework", "core")
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				config.use = append(config.use, path)
			}
		}
	}
	paths := c.Paths
	if c.script != nil {
		extra, err := c.script.classPath(c)
		if err != nil {
			return nil, err
		}
		paths = append(append([]string{}, paths...), extra...)
	}
	modules := append(append([]string{}, c.Modules...), resolved.Modules()...)
	args := append(runtime[1:], "-C", "-q", "-d", iniArg("include_path", includePath(config.use, modules, paths)), "-d", "date.timezone="+systemTimezone())
	for _, value := range config.extensions {
		args = append(args, "-d", iniArg("extension", value))
	}
	keys := make([]string, 0, len(config.args))
	for name := range config.args {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		for _, value := range config.args[name] {
			args = append(args, "-d", iniArg(name, value))
		}
	}

	// Append command arguments
	args = append(args, main)
	for _, arg := range resolved.Arguments(c.Args) {
		args = append(args, utf7.Encode(arg))
	}

	cmd := exec.Command(runtime[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	name := c.Command
	if _, ok := resolved.(*plugin); ok {
		name = "plugin"
	}
	cmd.Env, err = c.expandEnv(append(
		os.Environ(),
		"XP_EXE="+c.Exe,
		"XP_VERSION="+c.Version,
		"XP_MODEL="+c.Model.Name(),
		"XP_COMMAND="+name,
	))
	return cmd, err
}

func iniArg(name, value string) string {
	// PHP parses each -d value as INI, independently of shell quoting.
	return name + `="` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func locate(files ...string) (string, error) {
	for _, file := range files {
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			return filepath.Abs(file)
		}
		// Continue
	}

	return "", fmt.Errorf("Failed to locate %v", files)
}

func entry(dir, name string) (string, bool) {
	bin := filepath.Join(dir, "bin")
	files, err := os.ReadDir(bin)
	if err != nil {
		return "", false
	}
	for _, file := range files {
		if file.IsDir() || strings.HasSuffix(file.Name(), ".bat") {
			continue
		}
		if p, err := Plugin(name, dir, file.Name()); err == nil && p.commandName() == name {
			return filepath.Join(bin, file.Name()), true
		}
	}
	return "", false
}

func lookupCommand(c *commandline) (command, error) {
	if c.script != nil {
		return c.script, nil
	}
	if c.Command == "help" {
		return resolveHelp(c)
	}
	if c, ok := commands[c.Command]; ok {
		return c, nil
	}

	return findPlugin(c, c.Command)
}

func findPlugin(c *commandline, name string) (*plugin, error) {
	for _, dir := range append(append([]string{}, c.Modules...), ".") {
		if file, ok := entry(dir, name); ok {
			p, err := Plugin(name, dir, filepath.Base(file))
			if err == nil {
				p.local = true
				return p, nil
			}
		}
	}
	for _, dir := range c.vendors() {
		if entry, ok := entry(dir, name); ok {
			if c, err := Plugin(name, dir, filepath.Base(entry)); err == nil {
				return c, nil
			}
		}
	}

	return nil, fmt.Errorf("command not implemented: %s (run `xp list` to see available commands)", name)
}

func (c *commandline) vendors() []string {
	locations := composer.Locations()
	if c.project != nil {
		if cwd, err := os.Getwd(); err == nil && len(locations) > 0 {
			dir := c.project.VendorDir()
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(cwd, dir)
			}
			locations[0] = dir
		}
	}
	return locations
}

func includePath(use, modules, paths []string) string {
	separator := string(os.PathListSeparator)
	return strings.Join(append(append([]string{"."}, use...), modules...), separator) +
		separator + separator + strings.Join(append([]string{"."}, paths...), separator)
}

func exitStatus(err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status := exiterr.ExitCode(); status >= 0 {
			return status, nil
		}
	}

	return 255, err
}
