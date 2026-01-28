package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a JSON file and reloads the engine on changes.
type Watcher struct {
	filePath string
	engine   *Engine
	watcher  *fsnotify.Watcher
	onChange func(msg string) // Callback for logging
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewWatcher creates a new file watcher for hot-reload.
func NewWatcher(filePath string, engine *Engine) (*Watcher, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		filePath: absPath,
		engine:   engine,
		watcher:  fsWatcher,
		stop:     make(chan struct{}),
	}

	return w, nil
}

// SetOnChange sets the callback for reload notifications.
func (w *Watcher) SetOnChange(fn func(msg string)) {
	w.onChange = fn
}

// Start begins watching the file for changes.
func (w *Watcher) Start() error {
	// Watch the directory (more reliable for editors that do atomic saves)
	dir := filepath.Dir(w.filePath)
	if err := w.watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	w.wg.Add(1)
	go w.watch()

	return nil
}

// watch handles file system events.
func (w *Watcher) watch() {
	defer w.wg.Done()

	filename := filepath.Base(w.filePath)

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Check if the changed file is our target
			if filepath.Base(event.Name) != filename {
				continue
			}

			// Handle write or create events
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if err := w.reload(); err != nil {
					if w.onChange != nil {
						w.onChange(fmt.Sprintf("❌ Reload failed: %v", err))
					}
				} else {
					if w.onChange != nil {
						w.onChange("🔄 Data reloaded successfully")
					}
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			if w.onChange != nil {
				w.onChange(fmt.Sprintf("⚠️ Watcher error: %v", err))
			}

		case <-w.stop:
			return
		}
	}
}

// reload reads the file and updates the engine.
func (w *Watcher) reload() error {
	data, err := os.ReadFile(w.filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	w.engine.ReloadData(jsonData)
	return nil
}

// Stop stops the file watcher.
func (w *Watcher) Stop() error {
	close(w.stop)
	w.wg.Wait()
	return w.watcher.Close()
}
