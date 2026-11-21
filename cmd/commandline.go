package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/xp-runners/evolution/composer"
)

const ini = "xp.ini"

type commandline struct {
	Paths          []string
	Modules        []string
	Model          model
	Command        string
	Args           []string
	Exe            string
	Version        string
	home           string
	configFiles    []string
	explicitConfig bool
	envFiles       []string
	project        *composer.Definition
	script         *script
}

type model interface {
	Run(c *commandline) (int, error)
	fmt.Stringer
	Name() string
}

type command interface {
	Arguments(args []string) []string
	Modules() []string
}

var options = map[string]func(c *commandline, value string){
	"-cp": func(c *commandline, value string) {
		c.Paths = append(c.Paths, value)
	},
	"-cp?": func(c *commandline, value string) {
		c.Paths = append(c.Paths, "?"+value)
	},
	"-cp!": func(c *commandline, value string) {
		c.Paths = append(c.Paths, "!"+value)
	},
	"-m": func(c *commandline, value string) {
		c.Modules = append(c.Modules, value)
	},
	"-watch": func(c *commandline, value string) {
		c.Model = &runwatching{value}
	},
	"-repeat": func(c *commandline, value string) {
		c.Model = &runrepeatedly{Declaration: value}
	},
	"-env": func(c *commandline, value string) {
		c.envFiles = append(c.envFiles, value)
	},
	"-c": func(c *commandline, value string) {
		if info, err := os.Stat(value); err == nil && info.IsDir() {
			value = filepath.Join(value, ini)
		}
		c.configFiles = []string{value}
		c.explicitConfig = true
	},
}

var flags = map[string]func(c *commandline){
	"-n": func(c *commandline) {
		c.configFiles = nil
		c.explicitConfig = false
	},
	"-supervise": func(c *commandline) {
		c.Model = &supervise{}
	},
}

var shorthands = map[string]string{
	"-v": "version",
	"-e": "eval",
	"-d": "dump",
	"-w": "write",
	"-?": "help",
}

// Parse creates an XP runner command line from an array
func Parse(exe string, args []string, home, appdata, version string) (*commandline, error) {
	defaults := []string{ini}
	if home != "" {
		defaults = append(defaults, filepath.Join(home, ".xp", ini))
	}
	if appdata != "" {
		defaults = append(defaults, filepath.Join(appdata, "Xp", ini))
	}
	if bundled := filepath.Join(filepath.Dir(exe), ini); bundled != ini {
		defaults = append(defaults, bundled)
	}
	c := &commandline{
		Paths:   make([]string, 0),
		Modules: make([]string, 0),
		home:    home,
		Model:   &runonce{},
		Args:    make([]string, 0),
		Exe:     exe,
		Version: version,
	}
	if project, err := composer.File(composer.FileName); err == nil {
		c.project = project
		c.Paths = append(c.Paths, "?"+filepath.Join(project.VendorDir(), "autoload.php"))
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if err := c.parse(args, defaults, true); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commandline) parse(args, defaults []string, scripts bool) error {
	c.Command = "help"
	c.Args = []string{}
	c.Model = &runonce{}
	c.configFiles, c.explicitConfig = defaults, false
	for i := 0; i < len(args); i++ {
		if f, ok := options[args[i]]; ok {
			i++
			if i >= len(args) {
				return fmt.Errorf("Option %s requires a value", args[i-1])
			}

			f(c, args[i])
		} else if f, ok := flags[args[i]]; ok {
			f(c)
		} else if n, ok := shorthands[args[i]]; ok {
			c.Command = n
			c.Args = args[i+1:]
			break
		} else if strings.HasPrefix(args[i], "-") {
			return fmt.Errorf("Unknown option %v", args[i])
		} else {
			name := args[i]
			c.Args = args[i+1:]
			if isCommand(name) {
				c.Command = name
				if _, builtin := commands[name]; !builtin && name != "list" && scripts && c.project != nil {
					if body, ok := c.project.Scripts[name]; ok {
						tokens, err := splitArgs(body)
						if err != nil {
							return fmt.Errorf("Composer script %s: %w", name, err)
						}
						extra := c.Args
						if err := c.parse(tokens[1:], defaults, false); err != nil {
							return err
						}
						c.Args = append(append([]string{}, c.Args...), extra...)
					}
				}
			} else if info, err := os.Stat(name); err == nil && !info.IsDir() &&
				!strings.HasSuffix(name, ".class.php") && !strings.HasSuffix(name, ".xar") {
				c.Command, c.script = "script", &script{file: name}
			} else {
				c.Command, c.Args = "run", args[i:]
			}
			break
		}
	}

	if repeat, ok := c.Model.(*runrepeatedly); ok {
		var err error
		repeat.schedule, err = parseSchedule(repeat.Declaration)
		return err
	}
	return nil
}

func isCommand(name string) bool {
	return name != "" && strings.IndexFunc(name, func(r rune) bool {
		return r != '-' && !unicode.IsLower(r)
	}) == -1
}

// Runs this command line
func (c *commandline) Run() (int, error) {
	if c.Command == "list" {
		return c.list(os.Stdout)
	}
	return c.Model.Run(c)
}
