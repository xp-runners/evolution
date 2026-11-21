package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type runwatching struct {
	Directory string
}

func watchTree(watcher *fsnotify.Watcher, dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

func waitForChange(watcher *fsnotify.Watcher, running <-chan error) error {
	for {
		var changed bool
		var watchErr error
		select {
		case err := <-running:
			if err != nil {
				return err
			}
			running = nil
		case event, ok := <-watcher.Events:
			if !ok {
				watchErr = fmt.Errorf("watcher closed")
			} else if event.Has(fsnotify.Create) {
				// Keep new directories watched even if this event won't trigger a run.
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := watchTree(watcher, event.Name); err != nil && !os.IsNotExist(err) {
						watchErr = err
					}
				}
			}
			changed = event.Has(fsnotify.Write | fsnotify.Create | fsnotify.Rename | fsnotify.Remove)
		case err, ok := <-watcher.Errors:
			if !ok {
				watchErr = fmt.Errorf("watcher closed")
			} else {
				watchErr = err
			}
		}
		if watchErr != nil {
			// Don't leave PHP running if the watcher fails.
			if running != nil {
				if err := <-running; err != nil {
					return err
				}
			}
			return fmt.Errorf("could not watch: %w", watchErr)
		}
		if changed && running == nil {
			return nil
		}
	}
}

// Run runs the command, then watches for changes in the given directory
// and reruns it again.
func (r runwatching) Run(c *commandline) (int, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return 2, err
	}
	defer watcher.Close()
	dir, err := filepath.EvalSymlinks(r.Directory)
	if err != nil {
		return 2, fmt.Errorf("could not watch: %w", err)
	}
	if err := watchTree(watcher, dir); err != nil {
		return 2, fmt.Errorf("could not watch: %w", err)
	}
	for {
		// Consume notifications during the run so they don't trigger another run.
		running := make(chan error)
		go func() {
			_, err := (runonce{}).Run(c)
			running <- err
		}()
		if err := waitForChange(watcher, running); err != nil {
			return 2, err
		}
	}
}

func (r runwatching) Name() string {
	return "watch"
}

func (r runwatching) String() string {
	return "watch(" + r.Directory + ")"
}
