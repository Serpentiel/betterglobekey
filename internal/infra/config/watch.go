package config

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debounce coalesces bursts of file-system events (editors often write, rename,
// and chmod in quick succession) into a single reload.
const debounce = 100 * time.Millisecond

// Watch invokes onChange whenever the file at path changes, until ctx is done.
// It watches the parent directory, since editors frequently replace files
// atomically (write to a temp file, then rename), and debounces rapid events.
func Watch(ctx context.Context, path string, onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	defer func() { _ = watcher.Close() }()

	if err := watcher.Add(filepath.Dir(path)); err != nil {
		return err
	}

	target := filepath.Clean(path)

	var timer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if filepath.Clean(event.Name) != target {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			if timer != nil {
				timer.Stop()
			}

			timer = time.AfterFunc(debounce, onChange)
		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}
