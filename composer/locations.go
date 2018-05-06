package composer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var VendorDir = "vendor"

// See https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func xdgConfigDir() (string, bool) {
	if config, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return config, ok
	}

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "XDG_") {
			return filepath.Join(os.Getenv("HOME"), ".config"), true
		}
	}
	return "", false
}

// Locations returns composer locations for the current operating system
func Locations() []string {
	return LocationsOn(runtime.GOOS)
}

// LocationsFor returns composer locations on a given operating system
func LocationsOn(sys string) []string {
	locations := []string{}

	if wd, err := os.Getwd(); err == nil {
		locations = append(locations, filepath.Join(wd, VendorDir))
	}

	switch sys {
	case "windows":
		locations = append(locations, filepath.Join(os.Getenv("APPDATA"), "Composer", VendorDir))
		break

	case "darwin":
		locations = append(locations, filepath.Join(os.Getenv("HOME"), ".composer", VendorDir))
		break

	default:
		if dir, ok := xdgConfigDir(); ok {
			locations = append(locations, filepath.Join(dir, "composer", VendorDir))
		} else {
			locations = append(locations, filepath.Join(os.Getenv("HOME"), ".composer", VendorDir))
		}
	}

	return locations
}
