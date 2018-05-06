package composer

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func export(name, value string, run func()) {
	os.Setenv(name, value)
	run()
	os.Unsetenv(name)
}

func Test_locations_uses_runtime_os(t *testing.T) {
	assertEqual(LocationsOn(runtime.GOOS), Locations(), t)
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
	assertEqual(filepath.Join(os.Getenv("APPDATA"), "Composer", VendorDir), LocationsOn("windows")[1], t)
}

func Test_darwin_location(t *testing.T) {
	assertEqual(filepath.Join(os.Getenv("HOME"), ".composer", VendorDir), LocationsOn("darwin")[1], t)
}

func Test_unix_location(t *testing.T) {
	assertEqual(filepath.Join(os.Getenv("HOME"), ".composer", VendorDir), LocationsOn("unix")[1], t)
}

func Test_xdg_location_with_config_home(t *testing.T) {
	export("XDG_CONFIG_HOME", ".config", func() {
		assertEqual(filepath.Join(".config", "composer", VendorDir), LocationsOn("unix")[1], t)
	})
}

func Test_xdg_location(t *testing.T) {
	export("XDG_DATA_HOME", ".data", func() {
		assertEqual(filepath.Join(os.Getenv("HOME"), ".config", "composer", VendorDir), LocationsOn("unix")[1], t)
	})
}
