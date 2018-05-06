package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/xp-runners/evolution/composer"
)

type plugin struct {
	Dir     string
	Name    string
	Package string
	Class   string
}

func ucFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

// Plugin initializes a plugin
func Plugin(name, dir, file string) (*plugin, error) {
	spec := strings.Split(file, ".")

	switch len(spec) {
	case 3:
		return &plugin{Dir: dir, Name: spec[1], Package: spec[2], Class: "Runner"}, nil
	case 4:
		return &plugin{Dir: dir, Name: spec[1], Package: spec[2], Class: ucFirst(spec[3]) + "Runner"}, nil
	}

	return nil, fmt.Errorf("Malformed input string %v", file)
}

// Arguments formats command arguments for this command
func (r plugin) Arguments(args []string) []string {
	return append([]string{"xp." + r.Package + "." + r.Class}, args...)
}

// Modules returns additional modules to load for this command
func (r plugin) Modules() []string {
	modules := []string{filepath.Join(r.Dir, r.Name, r.Package)}

	for lib, _ := range composer.Dependencies(r.Dir, r.Name+"/"+r.Package) {
		modules = append(modules, filepath.Join(r.Dir, strings.Replace(lib, "/", string(os.PathSeparator), -1)))
	}

	return modules
}
