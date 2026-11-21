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

func waitForChange(watcher *fsnotify.Watcher) error {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher closed")
			}
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := watchTree(watcher, event.Name); err != nil && !os.IsNotExist(err) {
						return err
					}
				}
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
				return nil
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher closed")
			}
			return err
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
		if _, err := (runonce{}).Run(c); err != nil {
			return 2, err
		}

		if err := waitForChange(watcher); err != nil {
			return 2, fmt.Errorf("could not watch: %w", err)
		}
	}
}

func (r runwatching) Name() string {
	return "watch"
}

func (r runwatching) String() string {
	return "watch(" + r.Directory + ")"
}
