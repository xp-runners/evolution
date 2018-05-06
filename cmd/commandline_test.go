package cmd

import (
	"reflect"
	"testing"
)

const HOME = `/home/test`
const LOCALAPPDATA = `C:\Users\test\AppData\Local`
const VERSION = `9.0.0-dev`

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func mustParse(exe string, args []string) *commandline {
	cmd, err := Parse(exe, args, HOME, LOCALAPPDATA, VERSION)
	if err != nil {
		panic(err)
	}
	return cmd
}

func Test_run_is_default_command(t *testing.T) {
	assertEqual("run", mustParse("xp", []string{}).Command, t)
}

func Test_version_command(t *testing.T) {
	commands := []string{"run", "version", "eval", "web"}
	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			assertEqual(command, mustParse("xp", []string{command}).Command, t)
		})
	}
}

func Test_command_shorthands(t *testing.T) {
	shorthands := map[string]string{
		"-v": "version",
		"-e": "eval",
		"-d": "dump",
		"-w": "write",
		"-?": "help",
	}
	for shorthand, command := range shorthands {
		t.Run(shorthand, func(t *testing.T) {
			assertEqual(command, mustParse("xp", []string{shorthand}).Command, t)
		})
	}
}

func Test_classpath(t *testing.T) {
	cmd := mustParse("xp", []string{"-cp", "test.xar"})
	assertEqual([]string{"test.xar"}, cmd.Paths, t)
}

func Test_optional_classpath(t *testing.T) {
	cmd := mustParse("xp", []string{"-cp?", "test.xar"})
	assertEqual([]string{"?test.xar"}, cmd.Paths, t)
}

func Test_overlay_classpath(t *testing.T) {
	cmd := mustParse("xp", []string{"-cp!", "test.xar"})
	assertEqual([]string{"!test.xar"}, cmd.Paths, t)
}

func Test_multiple_classpaths(t *testing.T) {
	cmd := mustParse("xp", []string{"-cp", "test.xar", "-cp", "lib/"})
	assertEqual([]string{"test.xar", "lib/"}, cmd.Paths, t)
}

func Test_module(t *testing.T) {
	cmd := mustParse("xp", []string{"-m", "xp/unittest"})
	assertEqual([]string{"xp/unittest"}, cmd.Modules, t)
}

func Test_multiple_modules(t *testing.T) {
	cmd := mustParse("xp", []string{"-m", "xp/unittest", "-m", "xp/assertions"})
	assertEqual([]string{"xp/unittest", "xp/assertions"}, cmd.Modules, t)
}

func Test_runonce_is_default_model(t *testing.T) {
	cmd := mustParse("xp", []string{})
	assertEqual("once", cmd.Model.String(), t)
}

func Test_repeat_model(t *testing.T) {
	cmd := mustParse("xp", []string{"-repeat", "after 01:00"})
	assertEqual("repeat(after 01:00)", cmd.Model.String(), t)
}

func Test_watch_model(t *testing.T) {
	cmd := mustParse("xp", []string{"-watch", "src"})
	assertEqual("watch(src)", cmd.Model.String(), t)
}

func Test_classpath_without_value(t *testing.T) {
	_, err := Parse("xp", []string{"-cp"}, HOME, LOCALAPPDATA, VERSION)
	assertEqual("Option -cp requires a value", err.Error(), t)
}

func Test_unknown_option(t *testing.T) {
	_, err := Parse("xp", []string{"-x"}, HOME, LOCALAPPDATA, VERSION)
	assertEqual("Unknown option -x", err.Error(), t)
}
