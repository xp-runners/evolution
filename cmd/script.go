package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type script struct{ file string }

var scriptNamespace = regexp.MustCompile(`^(?:<\?php )?namespace\s+([^;]+);`)
var scriptDependency = regexp.MustCompile(`^use\s+.+\s+from\s+'([^'@]+)(?:@([^']+))?';`)

func (s script) Arguments(args []string) []string { return append([]string{s.file}, args...) }
func (s script) Modules() []string                { return nil }

func (s script) classPath(c *commandline) ([]string, error) {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return nil, err
	}
	var namespace string
	type dependency struct{ name, version string }
	var requires []dependency
	for _, line := range strings.Split(string(data), "\n") {
		if m := scriptNamespace.FindStringSubmatch(line); m != nil {
			namespace = strings.TrimSpace(m[1])
		} else if m := scriptDependency.FindStringSubmatch(line); m != nil {
			requires = append(requires, dependency{m[1], m[2]})
		}
	}
	if len(requires) == 0 {
		return nil, nil
	}
	var base string
	switch {
	case c.home == "" && os.Getenv("APPDATA") != "":
		base = filepath.Join(os.Getenv("APPDATA"), "Xp")
	case os.Getenv("XDG_CONFIG_HOME") != "":
		base = filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "xp")
	default:
		base = filepath.Join(c.home, ".xp")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "XDG_") {
				base = filepath.Join(c.home, ".config", "xp")
				break
			}
		}
	}
	local := filepath.Join(base, namespace)
	vendors := append([]string{filepath.Join(local, "vendor")}, c.vendors()...)
	var paths, missing []string
	seen := make(map[string]bool)
	for _, dep := range requires {
		found := false
		for _, dir := range vendors {
			if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(dep.name))); err == nil && info.IsDir() {
				loader := filepath.Join(dir, "autoload.php")
				if !seen[loader] {
					paths, seen[loader] = append(paths, loader), true
				}
				found = true
				break
			}
		}
		if !found {
			if dep.version != "" && dep.version != "*" {
				dep.name += ":" + dep.version
			}
			missing = append(missing, dep.name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%s: cannot find %s (searched %s; install in %s with Composer)", s.file,
			strings.Join(missing, ", "), strings.Join(vendors, ", "), local)
	}
	return paths, nil
}
