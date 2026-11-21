package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/xp-runners/evolution/composer"
)

type plugin struct {
	Dir          string
	Name         string
	Package      string
	Class        string
	local        bool
	explicitName string
}

func ucFirst(str string) string {
	if str == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(str)
	return string(unicode.ToUpper(r)) + str[n:]
}

// Plugin initializes a plugin
func Plugin(name, dir, file string) (*plugin, error) {
	spec := strings.Split(file, ".")
	if len(spec) < 3 || spec[0] != "xp" || spec[1] == "" || spec[2] == "" {
		return nil, fmt.Errorf("malformed plugin entry %q", file)
	}

	switch len(spec) {
	case 3:
		return &plugin{Dir: dir, Name: spec[1], Package: spec[2], Class: "Runner"}, nil
	case 4:
		if spec[3] != "" {
			return &plugin{Dir: dir, Name: spec[1], Package: spec[2], Class: ucFirst(spec[3]) + "Runner", explicitName: spec[3]}, nil
		}
	}

	return nil, fmt.Errorf("Malformed input string %v", file)
}

func (r plugin) commandName() string {
	if r.explicitName != "" {
		return r.explicitName
	}
	return r.Package
}

// Arguments formats command arguments for this command
func (r plugin) Arguments(args []string) []string {
	return append([]string{"xp." + r.Package + "." + r.Class}, args...)
}

// Modules returns additional modules to load for this command
func (r plugin) Modules() []string {
	if r.local {
		return nil
	}
	modules := []string{filepath.Join(r.Dir, r.Name, r.Package)}

	for _, lib := range composer.DependencyOrder(r.Dir, r.Name+"/"+r.Package) {
		modules = append(modules, filepath.Join(r.Dir, filepath.FromSlash(lib)))
	}

	return modules
}
