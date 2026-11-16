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

type ColorManager struct {
	mu             sync.RWMutex
	palette        *models.ColorPalette
	watcher        io.Closer
	themeJSONPath  string
	colorsJSONPath string
	notify         chan struct{}
}

// NewColorManager creates a new ColorManager with default paths.
func NewColorManager() (*ColorManager, error) {
	configDir, err := appPaths.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	themePath := DefaultDMSThemePath()
	colorsPath := filepath.Join(configDir, "colors.json")

	// Ensure the config directory exists for colorsJSONPath
	if err := os.MkdirAll(filepath.Dir(colorsPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return NewColorManagerWithPaths(themePath, colorsPath)
}

// NewColorManagerWithPaths creates a new ColorManager with custom paths.
// Precedence: themeJSONPath (if exists and valid JSON) > colorsJSONPath (if exists and valid JSON) > models.DefaultColorPalette().
// It only creates the colorsJSONPath parent directory (via appPaths.ConfigDir()), and does NOT overwrite themeJSONPath.
func NewColorManagerWithPaths(themeJSONPath, colorsJSONPath string) (*ColorManager, error) {
	cm := &ColorManager{
		palette:        models.DefaultColorPalette(),
		themeJSONPath:  themeJSONPath,
		colorsJSONPath: colorsJSONPath,
		notify:         make(chan struct{}, 1),
	}

	// Load with precedence: theme.json > colors.json > default
	if err := cm.loadWithPrecedence(); err != nil {
		// Already falls back to default in loadWithPrecedence, so we just log and continue
		_ = err
	}

	// Start watching both files
	if err := cm.startWatching(); err != nil {
		// Non-fatal; watcher will be nil and no watching will occur
		_ = err
	}

	return cm, nil
}

// loadWithPrecedence loads colors with precedence: theme.json > colors.json > default.
// If theme.json exists and is valid, use it. Otherwise, if colors.json exists and is valid, use it.
// Falls back to default palette if neither is valid.
func (cm *ColorManager) loadWithPrecedence() error {
	// Try theme.json first
	if _, err := os.Stat(cm.themeJSONPath); err == nil {
		if palette, err := cm.loadThemeJSON(); err == nil {
			cm.mu.Lock()
			cm.palette = palette
			cm.mu.Unlock()
			return nil
		}
		// theme.json exists but is invalid; continue to colors.json
	}

	// Try colors.json
	if _, err := os.Stat(cm.colorsJSONPath); err == nil {
		if palette, err := cm.loadColorsJSON(); err == nil {
			cm.mu.Lock()
			cm.palette = palette
			cm.mu.Unlock()
			return nil
		}
		// colors.json exists but is invalid; fall back to default
	}

	// Fall back to default palette when neither file is present and valid.
	cm.mu.Lock()
	cm.palette = models.DefaultColorPalette()
	cm.mu.Unlock()
	return nil
}

// loadThemeJSON loads and parses the theme.json file from dank-theme-engine.
func (cm *ColorManager) loadThemeJSON() (*models.ColorPalette, error) {
	data, err := os.ReadFile(cm.themeJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	doc, err := ParseDTEThemeDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme document: %w", err)
	}

	return doc.ToColorPalette(), nil
}

// loadColorsJSON loads the legacy colors.json file.
func (cm *ColorManager) loadColorsJSON() (*models.ColorPalette, error) {
	data, err := os.ReadFile(cm.colorsJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read colors file: %w", err)
	}

	var palette models.ColorPalette
	if err := json.Unmarshal(data, &palette); err != nil {
		return nil, fmt.Errorf("failed to unmarshal colors: %w", err)
	}

	return &palette, nil
}

func (cm *ColorManager) GetPalette() *models.ColorPalette {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	paletteCopy := *cm.palette
	return &paletteCopy
}

func (cm *ColorManager) Close() error {
	if cm.watcher != nil {
		return cm.watcher.Close()
	}
	return nil
}

func (cm *ColorManager) ColorChanges() <-chan struct{} {
	return cm.notify
}

func (cm *ColorManager) notifyColorChange() {
	select {
	case cm.notify <- struct{}{}:
	default:
	}
}

// startWatching sets up file watching for both theme.json and colors.json.
func (cm *ColorManager) startWatching() error {
	// Collect paths to watch
	var pathsToWatch []string

	// Always watch colorsJSONPath
	pathsToWatch = append(pathsToWatch, cm.colorsJSONPath)

	// Only watch themeJSONPath if its parent directory exists
	themeDir := filepath.Dir(cm.themeJSONPath)
	if _, err := os.Stat(themeDir); err == nil {
		pathsToWatch = append(pathsToWatch, cm.themeJSONPath)
	}

	watcher, err := watchConfigFiles(pathsToWatch, cm.onFileChanged)
	if err != nil {
		return err
	}

	cm.watcher = watcher
	return nil
}

// onFileChanged is called when any watched file changes.
// It reloads with precedence and notifies listeners.
func (cm *ColorManager) onFileChanged() {
	oldPalette := cm.GetPalette()

	if err := cm.loadWithPrecedence(); err != nil {
		// If loading fails, keep current palette
		// But still notify that something changed (the file might be corrupt now)
		cm.notifyColorChange()
		return
	}

	// Check if palette actually changed
	newPalette := cm.GetPalette()
	if paletteEqual(oldPalette, newPalette) {
		return
	}

	cm.notifyColorChange()
}

// paletteEqual checks if two palettes are equal.
func paletteEqual(a, b *models.ColorPalette) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Deep comparison
	return a.UI == b.UI &&
		a.Charts == b.Charts &&
		a.ProgressBars == b.ProgressBars &&
		a.Temperature == b.Temperature &&
		a.Status == b.Status
}
