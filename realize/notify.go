package realize

// this code is imported from moby, unfortunately i can't import it directly as dependencies from its repo,
// cause there was a problem between moby vendor and fsnotify
// i have just added only the walk methods and some little changes to polling interval, originally set as static.

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var (
	// errPollerClosed is returned when the poller is closed
	errPollerClosed = errors.New("poller is closed")
	// errNoSuchWatch is returned when trying to remove a watch that doesn't exist
	errNoSuchWatch = errors.New("watch does not exist")
)

type (
	// FileWatcher is an interface for implementing file notification watchers
	FileWatcher interface {
		Close() error
		Add(string) error
		Walk(string, bool) string
		Remove(string) error
		Errors() <-chan error
		Events() <-chan fsnotify.Event
	}
	// fsNotifyWatcher wraps the fsnotify package to satisfy the FileNotifier interface
	fsNotifyWatcher struct {
		*fsnotify.Watcher
	}
	// filePoller is used to poll files for changes, especially in cases where fsnotify
	// can't be run (e.g. when inotify handles are exhausted)
	// filePoller satisfies the FileWatcher interface
	filePoller struct {
		// watches is the list of files currently being polled, close the associated channel to stop the watch
		watches map[string]chan struct{}
		// events is the channel to listen to for watch events
		events chan fsnotify.Event
		// errors is the channel to listen to for watch errors
		errors chan error
		// mu locks the poller for modification
		mu sync.Mutex
		// closed is used to specify when the poller has already closed
		closed bool
		// polling interval
		interval time.Duration
	}
)

// PollingWatcher returns a poll-based file watcher
func PollingWatcher(interval time.Duration) FileWatcher {
	if interval == 0 {
		interval = time.Duration(1) * time.Second
	}
	return &filePoller{
		interval: interval,
		events:   make(chan fsnotify.Event),
		errors:   make(chan error),
	}
}

// NewFileWatcher tries to use an fs-event watcher, and falls back to the poller if there is an error
func NewFileWatcher(l Legacy) (FileWatcher, error) {
	if !l.Force {
		if w, err := EventWatcher(); err == nil {
			return w, nil
		}
	}
	return PollingWatcher(l.Interval), nil
}

// EventWatcher returns an fs-event based file watcher
func EventWatcher() (FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsNotifyWatcher{Watcher: w}, nil
}

// Errors returns the fsnotify error channel receiver
func (w *fsNotifyWatcher) Errors() <-chan error {
	return w.Watcher.Errors
}

// Events returns the fsnotify event channel receiver
func (w *fsNotifyWatcher) Events() <-chan fsnotify.Event {
	return w.Watcher.Events
}

// Walk fsnotify
func (w *fsNotifyWatcher) Walk(path string, init bool) string {
	if err := w.Add(path); err != nil {
		return ""
	}
	return path
}

// Close closes the poller
// All watches are stopped, removed, and the poller cannot be added to
func (w *filePoller) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}

	w.closed = true
	for name := range w.watches {
		w.remove(name)
		delete(w.watches, name)
	}
	w.mu.Unlock()
	return nil
}

// Errors returns the errors channel
// This is used for notifications about errors on watched files
func (w *filePoller) Errors() <-chan error {
	return w.errors
}

// Add adds a filename to the list of watches
// once added the file is polled for changes in a separate goroutine
func (w *filePoller) Add(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return errPollerClosed
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	fi, err := os.Stat(name)
	if err != nil {
		return err
	}

	if w.watches == nil {
		w.watches = make(map[string]chan struct{})
	}
	if _, exists := w.watches[name]; exists {
		return fmt.Errorf("watch exists")
	}
	chClose := make(chan struct{})
	w.watches[name] = chClose
	go w.watch(f, fi, chClose)
	return nil
}

// Remove poller
func (w *filePoller) remove(name string) error {
	if w.closed {
		return errPollerClosed
	}

	chClose, exists := w.watches[name]
	if !exists {
		return errNoSuchWatch
	}
	close(chClose)
	delete(w.watches, name)
	return nil
}

// Remove stops and removes watch with the specified name
func (w *filePoller) Remove(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.remove(name)
}

// Events returns the event channel
// This is used for notifications on events about watched files
func (w *filePoller) Events() <-chan fsnotify.Event {
	return w.events
}

// Walk poller
func (w *filePoller) Walk(path string, init bool) string {
	check := w.watches[path]
	if err := w.Add(path); err != nil {
		return ""
	}
	if check == nil && init {
		_, err := os.Stat(path)
		if err == nil {
			go w.sendEvent(fsnotify.Event{Op: fsnotify.Create, Name: path}, w.watches[path])
		}
	}
	return path
}

// sendErr publishes the specified error to the errors channel
func (w *filePoller) sendErr(e error, chClose <-chan struct{}) error {
	select {
	case w.errors <- e:
	case <-chClose:
		return fmt.Errorf("closed")
	}
	return nil
}

// sendEvent publishes the specified event to the events channel
func (w *filePoller) sendEvent(e fsnotify.Event, chClose <-chan struct{}) error {
	select {
	case w.events <- e:
	case <-chClose:
		return fmt.Errorf("closed")
	}
	return nil
}

// watch is responsible for polling the specified file for changes
// upon finding changes to a file or errors, sendEvent/sendErr is called
func (w *filePoller) watch(f *os.File, lastFi os.FileInfo, chClose chan struct{}) {
	defer f.Close()
	for {
		time.Sleep(w.interval)
		select {
		case <-chClose:
			logrus.Debugf("watch for %s closed", f.Name())
			return
		default:
		}

		fi, err := os.Stat(f.Name())
		switch {
		case err != nil && lastFi != nil:
			// If it doesn't exist at this point, it must have been removed
			// no need to send the error here since this is a valid operation
			if os.IsNotExist(err) {
				if err := w.sendEvent(fsnotify.Event{Op: fsnotify.Remove, Name: f.Name()}, chClose); err != nil {
					return
				}
				lastFi = nil
			}
			// at this point, send the error
			w.sendErr(err, chClose)
			return
		case lastFi == nil:
			if err := w.sendEvent(fsnotify.Event{Op: fsnotify.Create, Name: f.Name()}, chClose); err != nil {
				return
			}
			lastFi = fi
		case fi.Mode() != lastFi.Mode():
			if err := w.sendEvent(fsnotify.Event{Op: fsnotify.Chmod, Name: f.Name()}, chClose); err != nil {
				return
			}
			lastFi = fi
		case fi.ModTime() != lastFi.ModTime() || fi.Size() != lastFi.Size():
			if err := w.sendEvent(fsnotify.Event{Op: fsnotify.Write, Name: f.Name()}, chClose); err != nil {
				return
			}
			lastFi = fi
		}
	}
}
