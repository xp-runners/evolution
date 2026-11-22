package cmd

import (
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchIgnoresChangesDuringRun(t *testing.T) {
	events := make(chan fsnotify.Event)
	defer close(events)
	running := make(chan error)
	defer close(running)
	watcher := &fsnotify.Watcher{Events: events}
	result := make(chan error, 1)
	go func() { result <- waitForChange(watcher, running) }()

	deadline := time.After(5 * time.Second)
	send := func(op fsnotify.Op) {
		t.Helper()
		select {
		case events <- fsnotify.Event{Name: "source.php", Op: op}:
		case err := <-result:
			t.Fatalf("stopped waiting before the next change: %v", err)
		case <-deadline:
			t.Fatal("watcher did not consume the event")
		}
	}

	send(fsnotify.Create)
	send(fsnotify.Write)
	send(fsnotify.Rename)
	select {
	case running <- nil:
	case err := <-result:
		t.Fatalf("a change triggered a rerun while the process was running: %v", err)
	case <-deadline:
		t.Fatal("watcher did not observe process completion")
	}

	send(fsnotify.Write)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-deadline:
		t.Fatal("a change after process completion did not trigger a rerun")
	}
}
