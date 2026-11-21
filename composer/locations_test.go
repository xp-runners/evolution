package composer

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var NoFile = func(file string) bool { return false }
var FileExists = func(file string) bool { return true }

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func export(exports map[string]string, run func()) {
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "XDG_") {
			if i := strings.Index(v, "="); i >= 0 {
				os.Unsetenv(v[:i])
				defer os.Setenv(v[:i], v[i+1:])
			}
		}
	}

	for name, value := range exports {
		os.Setenv(name, value)
	}

	run()

	for name, _ := range exports {
		os.Unsetenv(name)
	}
}

func Test_locations_uses_runtime_os(t *testing.T) {
	assertEqual(LocationsOn(runtime.GOOS, StatDir), Locations(), t)
}

func Test_locations_starts_with_cwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Cannot get working directory: %v", err)
		return
	}

	assertEqual(filepath.Join(cwd, VendorDir), Locations()[0], t)
}

func Test_windows_location(t *testing.T) {
	assertEqual(filepath.Join(os.Getenv("APPDATA"), "Composer", VendorDir), LocationsOn("windows", NoFile)[1], t)
}

func Test_darwin_location(t *testing.T) {
	assertEqual(filepath.Join(os.Getenv("HOME"), ".composer", VendorDir), LocationsOn("darwin", NoFile)[1], t)
}

func Test_unix_location(t *testing.T) {
	assertEqual(filepath.Join(os.Getenv("HOME"), ".composer", VendorDir), LocationsOn("unix", FileExists)[1], t)
}

func Test_unix_location_fallback(t *testing.T) {
	export(map[string]string{}, func() {
		assertEqual(filepath.Join(os.Getenv("HOME"), ".composer", VendorDir), LocationsOn("unix", NoFile)[1], t)
	})
}

func Test_xdg_location_with_config_home(t *testing.T) {
	export(map[string]string{"XDG_CONFIG_HOME": ".config"}, func() {
		assertEqual(filepath.Join(".config", "composer", VendorDir), LocationsOn("unix", NoFile)[1], t)
	})
}

func Test_xdg_location(t *testing.T) {
	export(map[string]string{"XDG_DATA_HOME": ".data"}, func() {
		assertEqual(filepath.Join(os.Getenv("HOME"), ".config", "composer", VendorDir), LocationsOn("unix", NoFile)[1], t)
	})
}
