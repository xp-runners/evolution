package composer

import (
	"path/filepath"
	"sort"
	"strings"
)

// Dependencies returns all dependencies, including transitive ones, for a library
func Dependencies(dir, lib string) map[string]string {
	_, versions := dependencies(dir, lib)
	return versions
}

// DependencyOrder returns installed package names in deterministic traversal order.
func DependencyOrder(dir, lib string) []string {
	order, _ := dependencies(dir, lib)
	return order
}

func dependencies(dir, lib string) ([]string, map[string]string) {
	var order []string
	versions := make(map[string]string)
	seen := map[string]bool{lib: true}
	var walk func(string)
	walk = func(pkg string) {
		file, err := File(filepath.Join(dir, filepath.FromSlash(pkg), FileName))
		if err != nil {
			return
		}
		var packages []string
		for name := range file.Require {
			if strings.Contains(name, "/") {
				packages = append(packages, name)
			}
		}
		sort.Strings(packages)
		for _, name := range packages {
			if !seen[name] {
				seen[name] = true
				order = append(order, name)
				versions[name] = file.Require[name]
				walk(name)
			}
		}
	}
	walk(lib)
	return order, versions
}
