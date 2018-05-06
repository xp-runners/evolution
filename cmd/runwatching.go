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

func watch(dir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan error)
	go func() {
		for {
			select {
			case <-watcher.Events:
				done <- nil
				break
			case err := <-watcher.Errors:
				done <- err
				break
			}
		}
	}()

	// Add directory
	if err := watcher.Add(dir); err != nil {
		return err
	}

	// Add subdirectories
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name != "." && name != ".." {
				if err := watcher.Add(path); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return <-done
}

// Run runs the command, then watches for changes in the given directory
// and reruns it again.
func (r runwatching) Run(c *commandline) (int, error) {
	for {
		cmd, err := NewCmd(c)
		if err != nil {
			return 255, fmt.Errorf("Could not create process: %v", err)
		}

		if err := cmd.Start(); err != nil {
			return 255, fmt.Errorf("Could not start process: %v", err)
		}

		cmd.Wait()

		if err := watch(r.Directory); err != nil {
			return 255, fmt.Errorf("Could not watch: %v", err)
		}
	}
	return 0, nil
}

func (r runwatching) Name() string {
	return "watch"
}

func (r runwatching) String() string {
	return "watch(" + r.Directory + ")"
}
