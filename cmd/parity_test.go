package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func fixture(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

func useProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})
	return dir
}

func TestDispatch(t *testing.T) {
	useProject(t)
	fixture(t, "composer.json", `{
		"config":{"vendor-dir":"deps"},
		"scripts":{"check":"xp -n run Example.Cli \"two words\"","ignored":["php -v"]}
	}`)
	fixture(t, "tool.php", "<?php")
	for _, test := range []struct {
		input []string
		name  string
		args  []string
	}{
		{nil, "help", nil},
		{[]string{"Example.Cli", "-v"}, "run", []string{"Example.Cli", "-v"}},
		{[]string{"./example.xar"}, "run", []string{"./example.xar"}},
		{[]string{"tool.php", "arg"}, "script", []string{"arg"}},
		{[]string{"check", "extra"}, "run", []string{"Example.Cli", "two words", "extra"}},
		{[]string{"ignored"}, "ignored", []string{}},
	} {
		c := mustParse("xp", test.input)
		if c.Command != test.name || strings.Join(c.Args, "\x00") != strings.Join(test.args, "\x00") {
			t.Errorf("Parse(%q) = %s %q; want %s %q", test.input, c.Command, c.Args, test.name, test.args)
		}
		if want := "?" + filepath.Join("deps", "autoload.php"); c.Paths[0] != want {
			t.Errorf("classpath = %q; want %q", c.Paths, want)
		}
	}
}

func TestRuntimeConfigPrecedence(t *testing.T) {
	dir := useProject(t)
	t.Setenv("XP_RT", "selected")
	fixture(t, "xp.ini", "[runtime]\nflag=base\n[runtime@selected]\ndefault=selected-php\nflag='a b'\nflag=second")
	fixture(t, "home/.xp/xp.ini", "[runtime]\nflag=low\nlow=yes")
	c, err := Parse(filepath.Join(dir, "bin", "xp"), []string{"-v"}, filepath.Join(dir, "home"), "", VERSION)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.executable != "selected-php" || !reflect.DeepEqual(s.args["flag"], []string{"a b", "second"}) || s.args["low"][0] != "yes" {
		t.Fatalf("unexpected settings: %#v", s)
	}
	c = mustParse("xp", []string{"-c", "home/.xp", "-v"})
	s, err = c.settings()
	if err != nil || s.executable != "selected" || !reflect.DeepEqual(s.args["flag"], []string{"low"}) {
		t.Fatalf("unexpected explicit settings: %#v (%v)", s, err)
	}
}

func TestPluginDiscoveryAndHelp(t *testing.T) {
	dir := useProject(t)
	fixture(t, "composer.json", `{"config":{"vendor-dir":"deps"}}`)
	fixture(t, "deps/bin/xp.acme.tools.check", "#!/bin/sh\n")
	fixture(t, "deps/acme/tools/composer.json", `{"require":{"acme/extra":"*","ext-json":"*"}}`)
	fixture(t, "deps/acme/extra/composer.json", `{"require":{"acme/tools":"*"}}`)
	c := mustParse("xp", []string{"check"})
	p, err := lookupCommand(c)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(dir, "deps", "acme", "tools"), filepath.Join(dir, "deps", "acme", "extra")}
	for i := range want {
		want[i], _ = filepath.EvalSymlinks(want[i])
	}
	got := p.Modules()
	for i := range got {
		got[i], _ = filepath.EvalSymlinks(got[i])
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plugin modules %q; want %q", got, want)
	}
	c = mustParse("xp", []string{"help", "check/usage"})
	h, err := lookupCommand(c)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.Arguments(nil); !reflect.DeepEqual(got, []string{"xp.runtime.Help", "@xp/tools/check/usage.md"}) {
		t.Errorf("help argv = %q", got)
	}
	var out bytes.Buffer
	if status, err := c.list(&out); status != 0 || err != nil || !strings.Contains(out.String(), "$ xp check (from acme/tools)") {
		t.Fatalf("list: status=%d err=%v output=%q", status, err, out.String())
	}
}

func TestRepeatSchedule(t *testing.T) {
	start := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		spec          string
		initial, next time.Duration
	}{
		{"after 0:05", 0, 5 * time.Second},
		{"every 0:05", 0, 3 * time.Second},
		{"at 09:00 17:00", 4 * time.Hour, 4*time.Hour - 2*time.Second},
		{"at 09:00", 20 * time.Hour, 20*time.Hour - 2*time.Second},
	} {
		s, err := parseSchedule(test.spec + ",until 3|7")
		if err != nil {
			t.Fatal(err)
		}
		if s.initial(start) != test.initial || s.next(start, start.Add(2*time.Second)) != test.next || !s.until(7) || s.until(0) {
			t.Errorf("unexpected schedule for %s", test.spec)
		}
	}
	if _, err := parseSchedule("every nope"); err == nil {
		t.Fatal("invalid interval accepted")
	}
}

func TestPHPLaunch(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil {
		t.Skip("PHP CLI not installed")
	}
	if err := exec.Command(php, "-r", "exit(function_exists('iconv') ? 0 : 1);").Run(); err != nil {
		t.Skip("PHP iconv unavailable")
	}
	dir := useProject(t)
	t.Setenv("XP_RT", `"`+php+`"`)
	t.Setenv("XP_ENV", "test")
	t.Setenv("USE_XP", filepath.Join(dir, "core with spaces"))
	fixture(t, ".env", "PORT_VALUE=one\nCHAIN=$PORT_VALUE")
	fixture(t, ".env.test", "PORT_VALUE=two")
	fixture(t, ".env.local", `CHAIN="${CHAIN} ${PORT_VALUE} \$PORT_VALUE"`)
	fixture(t, "class-main.php", `<?php
		echo json_encode([
			get_include_path(), getenv('XP_MODEL'), getenv('XP_EXE'), getenv('CHAIN'),
			array_map(fn($v) => iconv('UTF-7', 'UTF-8', $v), array_slice($argv, 1))
		]);
		exit(7);
	`)
	c := mustParse(filepath.Join(dir, "xp"), []string{"-n", "Example.Main", "£-&", "", "two words"})
	process, err := NewCmd(c)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	status, err := exitStatus(process.Run())
	if err != nil || status != 7 {
		t.Fatalf("PHP exit = %d, %v: %s", status, err, stderr.String())
	}
	var got []any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("PHP output: %s: %v", stdout.String(), err)
	}
	want := []any{includePath([]string{filepath.Join(dir, "core with spaces")}, nil, nil),
		"default", c.Exe, "one two $PORT_VALUE", []string{"Example.Main", "£-&", "", "two words"}}
	w, _ := json.Marshal(want)
	var expected []any
	_ = json.Unmarshal(w, &expected)
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("PHP output = %s; want %s", stdout.String(), w)
	}
}

func TestWatchNewSubdirectory(t *testing.T) {
	dir := t.TempDir()
	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err := watchTree(w, dir); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	wait := func() {
		t.Helper()
		result := make(chan error, 1)
		go func() { result <- waitForChange(w) }()
		select {
		case err := <-result:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("no change event")
		}
	}
	wait()
	fixture(t, filepath.Join(sub, "change.txt"), "changed")
	wait()
}

// A child test process provides a platform-independent supervision fixture.
func TestSupervisedFixture(t *testing.T) {
	if os.Getenv("XP_TEST_SUPERVISION") != "1" {
		return
	}
	conn, err := net.Dial("tcp4", "127.0.0.1:"+os.Getenv("XP_SIGNAL"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := os.WriteFile(os.Getenv("XP_TEST_READY"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	var command string
	if _, err := fmt.Fscan(conn, &command); err != nil || command != "QUIT" {
		t.Fatalf("shutdown command = %q (%v)", command, err)
	}
	if err := os.WriteFile(os.Getenv("XP_TEST_READY")+".done", nil, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestSuperviseGracefulShutdown(t *testing.T) {
	stop := make(chan struct{})
	defer func() {
		select {
		case <-stop:
		default:
			close(stop)
		}
	}()
	ready := filepath.Join(t.TempDir(), "ready")
	done := make(chan error, 1)
	go func() {
		status, _, err := superviseChild(stop, func() (*exec.Cmd, error) {
			child := exec.Command(os.Args[0], "-test.run=^TestSupervisedFixture$")
			child.Env = append(os.Environ(), "XP_TEST_SUPERVISION=1", "XP_TEST_READY="+ready)
			child.Stdout, child.Stderr = os.Stdout, os.Stderr
			return child, nil
		})
		if err == nil && status != 0 {
			err = fmt.Errorf("supervised child exited with %d", status)
		}
		done <- err
	}()
	deadline := time.After(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("supervisor stopped before child readiness: %v", err)
		case <-deadline:
			t.Fatal("supervised child did not connect")
		case <-time.After(10 * time.Millisecond):
		}
	}
	close(stop)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("supervisor did not stop")
	}
	if _, err := os.Stat(ready + ".done"); err != nil {
		t.Fatalf("child did not receive QUIT: %v", err)
	}
}
