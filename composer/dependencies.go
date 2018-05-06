package composer

import "path/filepath"

// Dependencies returns all dependencies, including transitive ones, for a library
func Dependencies(dir, lib string) map[string]string {
	r := map[string]string{}

	file, err := File(filepath.Join(dir, lib, FileName))
	if err != nil {
		return r
	}

	for dependency, version := range file.Require {
		if "php" == dependency {
			continue
		} else if _, ok := r[dependency]; ok {
			continue
		}

		r[dependency] = version
		for transitive, version := range Dependencies(dir, dependency) {
			r[transitive] = version
		}
	}
	return r
}
