package config

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type watcherCloser struct {
	watcher  *fsnotify.Watcher
	done     chan struct{}
	waitDone sync.WaitGroup
}

// Close stops the watcher goroutine and closes the underlying fsnotify watcher.
func (wc *watcherCloser) Close() error {
	close(wc.done)
	var err error
	if wc.watcher != nil {
		err = wc.watcher.Close()
	}
	wc.waitDone.Wait()
	return err
}

// watchConfigFiles watches the parent directory of each target path and invokes
// onChange (debounced by 50ms) when any watched target receives Write, Create,
// Rename, or Remove events. Missing parent directories are skipped gracefully.
func watchConfigFiles(targetPaths []string, onChange func()) (io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	targetSet := make(map[string]bool, len(targetPaths))
	watchedDirs := make(map[string]bool, len(targetPaths))

	for _, targetPath := range targetPaths {
		cleanedPath := filepath.Clean(targetPath)
		targetSet[cleanedPath] = true

		dir := filepath.Dir(cleanedPath)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if watchedDirs[dir] {
			continue
		}
		if err := watcher.Add(dir); err != nil {
			continue
		}
		watchedDirs[dir] = true
	}

	wc := &watcherCloser{
		watcher: watcher,
		done:    make(chan struct{}),
	}

	wc.waitDone.Add(1)
	go func() {
		defer wc.waitDone.Done()

		var pendingTimer *time.Timer
		var timerC <-chan time.Time

		stopTimer := func() {
			if pendingTimer != nil {
				if !pendingTimer.Stop() {
					select {
					case <-pendingTimer.C:
					default:
					}
				}
				pendingTimer = nil
			}
			timerC = nil
		}

		for {
			select {
			case <-wc.done:
				stopTimer()
				return

			case <-timerC:
				pendingTimer = nil
				timerC = nil
				onChange()

			case event, ok := <-watcher.Events:
				if !ok {
					stopTimer()
					return
				}

				if !targetSet[filepath.Clean(event.Name)] {
					continue
				}

				if !event.Has(fsnotify.Write) &&
					!event.Has(fsnotify.Create) &&
					!event.Has(fsnotify.Rename) &&
					!event.Has(fsnotify.Remove) {
					continue
				}

				if pendingTimer == nil {
					pendingTimer = time.NewTimer(50 * time.Millisecond)
					timerC = pendingTimer.C
				} else {
					pendingTimer.Reset(50 * time.Millisecond)
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					stopTimer()
					return
				}
			}
		}
	}()

	return wc, nil
}
