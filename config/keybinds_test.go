package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JonasNowak/dms-gop/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeybindManager_Default(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()

	// Create a temp file path
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	km := &KeybindManager{
		binds:    models.DefaultKeybinds(),
		filePath: keybindsPath,
		notify:   make(chan struct{}, 1),
	}

	// Test Resolve returns defaults
	resolved := km.Resolve()
	assert.NotNil(t, resolved)
	assert.Equal(t, models.ActionQuit, resolved["q"])
	assert.Equal(t, models.ActionQuit, resolved["ctrl+c"])

	// Test PrimaryKey
	assert.Equal(t, "q", km.PrimaryKey(models.ActionQuit))
	assert.Equal(t, "r", km.PrimaryKey(models.ActionRefresh))

	// Cleanup
	_ = km.Close()
}

func TestKeybindManager_LoadConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	// Create a custom keybinds file
	customBinds := models.Keybinds{
		models.ActionQuit: []string{"ctrl+d"},
	}
	data, err := json.MarshalIndent(customBinds, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keybindsPath, data, 0644))

	// Create manager (we can't use NewKeybindManager since it uses appPaths)
	km := &KeybindManager{
		binds:    models.DefaultKeybinds(),
		filePath: keybindsPath,
		notify:   make(chan struct{}, 1),
	}

	// Load the config
	err = km.loadConfigFile()
	require.NoError(t, err)

	// Test that custom binds are loaded
	resolved := km.Resolve()
	assert.Equal(t, models.ActionQuit, resolved["ctrl+d"])
	// Default should still be there
	assert.Equal(t, models.ActionQuit, resolved["q"])

	// Test PrimaryKey returns custom key
	assert.Equal(t, "ctrl+d", km.PrimaryKey(models.ActionQuit))

	// Cleanup
	_ = km.Close()
}

func TestKeybindManager_CorruptConfig_FallsBackToDefault(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	// Create a corrupt keybinds file
	require.NoError(t, os.WriteFile(keybindsPath, []byte("{invalid json"), 0644))

	// Create manager
	km := &KeybindManager{
		binds:    models.DefaultKeybinds(),
		filePath: keybindsPath,
		notify:   make(chan struct{}, 1),
	}

	// Load should fail, but we handle it gracefully
	err := km.loadConfigFile()
	assert.Error(t, err)

	// binds should still be default
	resolved := km.Resolve()
	assert.Equal(t, models.ActionQuit, resolved["q"])

	// Cleanup
	_ = km.Close()
}

func TestKeybindManager_LiveReload(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	// Create initial keybinds file
	initialBinds := models.Keybinds{
		models.ActionQuit: []string{"ctrl+d"},
	}
	data, err := json.MarshalIndent(initialBinds, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keybindsPath, data, 0644))

	// Create manager using real constructor
	km, err := NewKeybindManagerWithPath(keybindsPath)
	require.NoError(t, err)
	defer km.Close()

	// Verify initial state
	assert.Equal(t, "ctrl+d", km.PrimaryKey(models.ActionQuit))

	changes := km.Changes()

	// Modify the file using atomic write + rename
	tmpFile := filepath.Join(tmpDir, "keybinds.json.tmp")
	newBinds := models.Keybinds{
		models.ActionQuit: []string{"ctrl+x"},
	}
	newData, err := json.MarshalIndent(newBinds, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile, newData, 0644))
	require.NoError(t, os.Rename(tmpFile, keybindsPath))

	// Wait for change notification
	select {
	case <-changes:
		// Change detected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for change notification")
	}

	// Verify the binds were updated
	assert.Equal(t, "ctrl+x", km.PrimaryKey(models.ActionQuit))
}

func TestKeybindManager_ConsecutiveLiveReloads(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	// Create initial keybinds file
	initialBinds := models.Keybinds{
		models.ActionQuit: []string{"ctrl+d"},
	}
	data, err := json.MarshalIndent(initialBinds, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keybindsPath, data, 0644))

	// Create manager using real constructor
	km, err := NewKeybindManagerWithPath(keybindsPath)
	require.NoError(t, err)
	defer km.Close()

	changes := km.Changes()

	// Do first atomic write + rename
	tmpFile1 := filepath.Join(tmpDir, "keybinds.json.tmp1")
	newBinds1 := models.Keybinds{
		models.ActionQuit: []string{"ctrl+x"},
	}
	newData1, err := json.MarshalIndent(newBinds1, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile1, newData1, 0644))
	require.NoError(t, os.Rename(tmpFile1, keybindsPath))

	// Wait for first change
	select {
	case <-changes:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for first change notification")
	}
	assert.Equal(t, "ctrl+x", km.PrimaryKey(models.ActionQuit))

	// Do second atomic write + rename
	tmpFile2 := filepath.Join(tmpDir, "keybinds.json.tmp2")
	newBinds2 := models.Keybinds{
		models.ActionQuit: []string{"ctrl+z"},
	}
	newData2, err := json.MarshalIndent(newBinds2, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile2, newData2, 0644))
	require.NoError(t, os.Rename(tmpFile2, keybindsPath))

	// Wait for second change
	select {
	case <-changes:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for second change notification")
	}
	assert.Equal(t, "ctrl+z", km.PrimaryKey(models.ActionQuit))

	// Do third atomic write + rename
	tmpFile3 := filepath.Join(tmpDir, "keybinds.json.tmp3")
	newBinds3 := models.Keybinds{
		models.ActionQuit: []string{"ctrl+c"},
	}
	newData3, err := json.MarshalIndent(newBinds3, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile3, newData3, 0644))
	require.NoError(t, os.Rename(tmpFile3, keybindsPath))

	// Wait for third change
	select {
	case <-changes:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for third change notification")
	}
	assert.Equal(t, "ctrl+c", km.PrimaryKey(models.ActionQuit))
}

func TestKeybindManager_LiveReload_CorruptFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	keybindsPath := filepath.Join(tmpDir, "keybinds.json")

	// Create initial valid keybinds file
	initialBinds := models.Keybinds{
		models.ActionQuit: []string{"ctrl+d"},
	}
	data, err := json.MarshalIndent(initialBinds, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keybindsPath, data, 0644))

	// Create manager using real constructor
	km, err := NewKeybindManagerWithPath(keybindsPath)
	require.NoError(t, err)
	defer km.Close()

	changes := km.Changes()

	// Make the file corrupt
	tmpFile := filepath.Join(tmpDir, "keybinds.json.tmp")
	require.NoError(t, os.WriteFile(tmpFile, []byte("{invalid json"), 0644))
	require.NoError(t, os.Rename(tmpFile, keybindsPath))

	// Wait for change notification
	select {
	case <-changes:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for change notification on corrupt file")
	}

	// Should fall back to defaults
	assert.Equal(t, "q", km.PrimaryKey(models.ActionQuit))
}
