package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/JonasNowak/dms-gop/models"
)

type KeybindManager struct {
	mu       sync.RWMutex
	binds    models.Keybinds
	watcher  io.Closer
	filePath string
	notify   chan struct{}
}

func NewKeybindManager() (*KeybindManager, error) {
	configDir, err := appPaths.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return NewKeybindManagerWithPath(filepath.Join(configDir, "keybinds.json"))
}

// NewKeybindManagerWithPath initializes a KeybindManager backed by filePath.
func NewKeybindManagerWithPath(filePath string) (*KeybindManager, error) {
	km := &KeybindManager{
		binds:    models.DefaultKeybinds(),
		filePath: filePath,
		notify:   make(chan struct{}, 1),
	}

	if err := km.loadOrCreateConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize keybinds file: %w", err)
	}

	if err := km.startWatching(); err != nil {
		return nil, fmt.Errorf("failed to start file watching: %w", err)
	}

	return km, nil
}

// Resolve returns a key→action lookup, merging user binds on top of the
// built-in defaults so defaults always keep working (additive).
func (km *KeybindManager) Resolve() map[string]models.KeyAction {
	km.mu.RLock()
	defer km.mu.RUnlock()

	out := make(map[string]models.KeyAction)
	for action, keys := range models.DefaultKeybinds() {
		for _, k := range keys {
			out[k] = action
		}
	}
	for action, keys := range km.binds {
		for _, k := range keys {
			out[k] = action
		}
	}
	return out
}

// PrimaryKey returns the first key bound to an action for display in help
// text, falling back to the built-in default when the user hasn't set one.
func (km *KeybindManager) PrimaryKey(action models.KeyAction) string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if keys := km.binds[action]; len(keys) > 0 {
		return keys[0]
	}
	if keys := models.DefaultKeybinds()[action]; len(keys) > 0 {
		return keys[0]
	}
	return ""
}

func (km *KeybindManager) Close() error {
	if km.watcher != nil {
		return km.watcher.Close()
	}
	return nil
}

func (km *KeybindManager) Changes() <-chan struct{} {
	return km.notify
}

func (km *KeybindManager) notifyChange() {
	select {
	case km.notify <- struct{}{}:
	default:
	}
}

func (km *KeybindManager) loadOrCreateConfigFile() error {
	if _, err := os.Stat(km.filePath); os.IsNotExist(err) {
		return km.createDefaultConfigFile()
	}
	return km.loadConfigFile()
}

func (km *KeybindManager) createDefaultConfigFile() error {
	data, err := json.MarshalIndent(km.binds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default keybinds: %w", err)
	}
	if err := os.WriteFile(km.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write keybinds file: %w", err)
	}
	return nil
}

func (km *KeybindManager) loadConfigFile() error {
	data, err := os.ReadFile(km.filePath)
	if err != nil {
		return fmt.Errorf("failed to read keybinds file: %w", err)
	}

	var binds models.Keybinds
	if err := json.Unmarshal(data, &binds); err != nil {
		return err
	}

	km.mu.Lock()
	km.binds = binds
	km.mu.Unlock()

	return nil
}

// reload reloads keybinds from disk, falling back to defaults on error, and notifies listeners.
func (km *KeybindManager) reload() {
	if err := km.loadConfigFile(); err != nil {
		km.mu.Lock()
		km.binds = models.DefaultKeybinds()
		km.mu.Unlock()
	}
	km.notifyChange()
}

// startWatching sets up file watching using the shared watchConfigFiles helper.
func (km *KeybindManager) startWatching() error {
	watcher, err := watchConfigFiles([]string{km.filePath}, km.reload)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	km.watcher = watcher
	return nil
}
