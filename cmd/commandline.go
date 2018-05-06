package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

const ini = "xp.ini"

type commandline struct {
	Paths   []string
	Modules []string
	Model   model
	Config  config
	Command string
	Args    []string
	Exe     string
	Version string
}

type model interface {
	Run(c *commandline) (int, error)
	fmt.Stringer
}

type config interface {
	Use() *string
	Exe() *string
	Extensions() []string
	Arguments() map[string]string
	fmt.Stringer
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
		c.Model = &runrepeatedly{value}
	},
	"-c": func(c *commandline, value string) {
		c.Config = ConfigFile(value)
	},
}

var flags = map[string]func(c *commandline){
	"-n": func(c *commandline) {
		c.Config = Environment()
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
	c := &commandline{
		Paths:   make([]string, 0),
		Modules: make([]string, 0),
		Command: "run",
		Config: Configs(
			Environment(),
			ConfigFile(ini),
			ConfigFile(filepath.Join(home, ini)),
			ConfigFile(filepath.Join(home, ".xp", ini)),
			ConfigFile(filepath.Join(appdata, "Xp", ini)),
			ConfigFile(filepath.Join(filepath.Dir(exe), ini)),
		),
		Model:   &runonce{},
		Args:    make([]string, 0),
		Exe:     exe,
		Version: version,
	}

	for i := 0; i < len(args); i++ {
		if f, ok := options[args[i]]; ok {
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("Option %s requires a value", args[i-1])
			}

			f(c, args[i])
		} else if f, ok := flags[args[i]]; ok {
			f(c)
		} else if n, ok := shorthands[args[i]]; ok {
			c.Command = n
			c.Args = args[i+1:]
			break
		} else if strings.HasPrefix(args[i], "-") {
			return nil, fmt.Errorf("Unknown option %v", args[i])
		} else {
			c.Command = args[i]
			c.Args = args[i+1:]
			break
		}
	}

	return c, nil
}

// Runs this command line
func (c *commandline) Run() (int, error) {
	return c.Model.Run(c)
}
