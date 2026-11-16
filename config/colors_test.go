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

func TestParseDTEThemeDocument_M3ColorsWithFloatEpoch(t *testing.T) {
	// Test with M3Color objects and float epoch generatedAt
	jsonData := `{
		"themeName": "Test Theme",
		"isDark": true,
		"palette": {
			"primary": {"hex": "#FF5733"},
			"onPrimary": {"hex": "#FFFFFF"},
			"primaryContainer": {"hex": "#FF7A50"},
			"onPrimaryContainer": {"hex": "#000000"},
			"inversePrimary": {"hex": "#33FF57"},
			"secondary": {"hex": "#33FF57"},
			"onSecondary": {"hex": "#FFFFFF"},
			"surface": {"hex": "#121212"},
			"onSurface": {"hex": "#FFFFFF"},
			"surfaceVariant": {"hex": "#1E1E1E"},
			"onSurfaceVariant": {"hex": "#CCCCCC"},
			"outline": {"hex": "#888888"},
			"error": {"hex": "#FF0000"},
			"onError": {"hex": "#FFFFFF"},
			"info": {"hex": "#0000FF"},
			"onInfo": {"hex": "#FFFFFF"},
			"warning": {"hex": "#FFCC00"},
			"onWarning": {"hex": "#000000"},
			"success": {"hex": "#00FF00"},
			"onSuccess": {"hex": "#000000"}
		},
		"dank16": {
			"colors": [
				{"hex": "#000000"},
				{"hex": "#FF0000"},
				{"hex": "#00FF00"},
				{"hex": "#FFFF00"},
				{"hex": "#0000FF"},
				{"hex": "#FF00FF"},
				{"hex": "#00FFFF"},
				{"hex": "#AAAAAA"},
				{"hex": "#555555"},
				{"hex": "#FF4444"}
			],
			"background": {"hex": "#000000"},
			"foreground": {"hex": "#FFFFFF"},
			"cursor": {"hex": "#FF0000"}
		},
		"generatedAt": 1794818100.0
	}`

	doc, err := ParseDTEThemeDocument([]byte(jsonData))
	require.NoError(t, err)

	palette := doc.ToColorPalette()

	// Verify M3 color mappings
	assert.Equal(t, "#FF5733", palette.UI.BorderPrimary)
	assert.Equal(t, "#33FF57", palette.UI.BorderSecondary)
	assert.Equal(t, "#FF7A50", palette.UI.HeaderBackground)
	assert.Equal(t, "#000000", palette.UI.HeaderText)
	assert.Equal(t, "#FFFFFF", palette.UI.TextPrimary)
	assert.Equal(t, "#CCCCCC", palette.UI.FooterText)

	// Verify dank16 mappings
	assert.Equal(t, "#00FFFF", palette.Charts.NetworkDownload) // dank16[6]
	assert.Equal(t, "#FF00FF", palette.Charts.NetworkUpload)   // dank16[5]
	assert.Equal(t, "#888888", palette.Charts.NetworkLine)     // outline

	// Verify temperature mappings
	assert.Equal(t, "#0000FF", palette.Temperature.Cold)   // dank16[4]
	assert.Equal(t, "#FFFF00", palette.Temperature.Warm)   // dank16[3]
	assert.Equal(t, "#FF0000", palette.Temperature.Hot)    // dank16[1]
	assert.Equal(t, "#FF4444", palette.Temperature.Danger) // dank16[9]

	// Verify status colors
	assert.Equal(t, "#00FF00", palette.Status.Success)
	assert.Equal(t, "#FFCC00", palette.Status.Warning)
	assert.Equal(t, "#FF0000", palette.Status.Error)
	assert.Equal(t, "#0000FF", palette.Status.Info)

	// Verify generatedAt was parsed correctly
	assert.Equal(t, float64(1794818100.0), doc.GeneratedAt)
}

func TestParseDTEThemeDocument_BareHexStringsWithRFC3339Epoch(t *testing.T) {
	// Test with bare hex strings and RFC3339 generatedAt
	jsonData := `{
		"themeName": "Test Theme 2",
		"isDark": false,
		"palette": {
			"primary": "#123456",
			"onPrimary": "#FFFFFF",
			"primaryContainer": "#ABCDEF",
			"onPrimaryContainer": "#000000",
			"inversePrimary": "#FEDCBA",
			"surface": "#FFFFFF",
			"onSurface": "#000000",
			"surfaceVariant": "#EEEEEE",
			"onSurfaceVariant": "#111111"
		},
		"dank16": {
			"colors": ["#000000", "#111111", "#222222", "#333333", "#444444", "#555555", "#666666", "#777777"],
			"background": "#000000",
			"foreground": "#FFFFFF",
			"cursor": "#FF0000"
		},
		"generatedAt": "2024-01-15T10:30:00Z"
	}`

	doc, err := ParseDTEThemeDocument([]byte(jsonData))
	require.NoError(t, err)

	palette := doc.ToColorPalette()

	// Verify bare hex string mappings work
	assert.Equal(t, "#123456", palette.UI.BorderPrimary)
	assert.Equal(t, "#FEDCBA", palette.UI.BorderSecondary)
	assert.Equal(t, "#ABCDEF", palette.UI.HeaderBackground)
	assert.Equal(t, "#000000", palette.UI.HeaderText)

	// Verify generatedAt was parsed from RFC3339
	assert.NotZero(t, doc.GeneratedAt)
}

func TestParseDTEThemeDocument_CorruptJSON(t *testing.T) {
	// Test that corrupt JSON returns an error
	jsonData := `{invalid json`

	_, err := ParseDTEThemeDocument([]byte(jsonData))
	assert.Error(t, err)
}

func TestColorManager_Precedence_ThemeOverColors(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a theme.json file
	themePath := filepath.Join(tmpDir, "theme.json")
	themeData := `{
		"themeName": "Test",
		"isDark": true,
		"palette": {"primary": {"hex": "#THEMEPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(themePath, []byte(themeData), 0644))

	// Create a colors.json file
	colorsPath := filepath.Join(tmpDir, "colors.json")
	colorsData := `{
		"ui": {"border_primary": "#COLORSPRIMARY"},
		"charts": {},
		"progress_bars": {},
		"temperature": {},
		"status": {}
	}`
	require.NoError(t, os.WriteFile(colorsPath, []byte(colorsData), 0644))

	// Create color manager with both paths
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Theme should take precedence
	palette := cm.GetPalette()
	assert.Equal(t, "#THEMEPRIMARY", palette.UI.BorderPrimary)
}

func TestColorManager_Precedence_ColorsOverDefault(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Don't create theme.json (it doesn't exist)
	themePath := filepath.Join(tmpDir, "theme.json")

	// Create a colors.json file
	colorsPath := filepath.Join(tmpDir, "colors.json")
	colorsData := `{
		"ui": {"border_primary": "#COLORSPRIMARY"},
		"charts": {},
		"progress_bars": {},
		"temperature": {},
		"status": {}
	}`
	require.NoError(t, os.WriteFile(colorsPath, []byte(colorsData), 0644))

	// Create color manager with both paths
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Colors should be used since theme doesn't exist
	palette := cm.GetPalette()
	assert.Equal(t, "#COLORSPRIMARY", palette.UI.BorderPrimary)
}

func TestColorManager_Precedence_DefaultFallback(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Don't create any config files
	themePath := filepath.Join(tmpDir, "theme.json")
	colorsPath := filepath.Join(tmpDir, "colors.json")

	// Create color manager with both paths
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Should fall back to default palette
	palette := cm.GetPalette()
	defaultPalette := models.DefaultColorPalette()
	assert.Equal(t, defaultPalette.UI.BorderPrimary, palette.UI.BorderPrimary)
}

func TestColorManager_Precedence_CorruptTheme_FallsBackToColors(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a corrupt theme.json file
	themePath := filepath.Join(tmpDir, "theme.json")
	require.NoError(t, os.WriteFile(themePath, []byte("{invalid json"), 0644))

	// Create a valid colors.json file
	colorsPath := filepath.Join(tmpDir, "colors.json")
	colorsData := `{
		"ui": {"border_primary": "#COLORSPRIMARY"},
		"charts": {},
		"progress_bars": {},
		"temperature": {},
		"status": {}
	}`
	require.NoError(t, os.WriteFile(colorsPath, []byte(colorsData), 0644))

	// Create color manager with both paths
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Should fall back to colors.json since theme.json is corrupt
	palette := cm.GetPalette()
	assert.Equal(t, "#COLORSPRIMARY", palette.UI.BorderPrimary)
}

func TestColorManager_Precedence_CorruptColors_FallsBackToTheme(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a valid theme.json file
	themePath := filepath.Join(tmpDir, "theme.json")
	themeData := `{
		"themeName": "Test",
		"isDark": true,
		"palette": {"primary": {"hex": "#THEMEPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(themePath, []byte(themeData), 0644))

	// Create a corrupt colors.json file
	colorsPath := filepath.Join(tmpDir, "colors.json")
	require.NoError(t, os.WriteFile(colorsPath, []byte("{invalid json"), 0644))

	// Create color manager with both paths
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Theme should be used since colors.json is corrupt
	palette := cm.GetPalette()
	assert.Equal(t, "#THEMEPRIMARY", palette.UI.BorderPrimary)
}

func TestColorManager_LiveReload(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create initial theme.json file
	themePath := filepath.Join(tmpDir, "theme.json")
	themeData := `{
		"themeName": "Test",
		"isDark": true,
		"palette": {"primary": {"hex": "#THEMEPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(themePath, []byte(themeData), 0644))

	colorsPath := filepath.Join(tmpDir, "colors.json")

	// Create color manager
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Get initial palette
	palette := cm.GetPalette()
	assert.Equal(t, "#THEMEPRIMARY", palette.UI.BorderPrimary)

	// Set up channel to receive change notifications
	changes := cm.ColorChanges()

	// Modify the theme.json file using atomic write + rename
	tmpFile := filepath.Join(tmpDir, "theme.json.tmp")
	newThemeData := `{
		"themeName": "Test2",
		"isDark": true,
		"palette": {"primary": {"hex": "#NEWPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(tmpFile, []byte(newThemeData), 0644))
	require.NoError(t, os.Rename(tmpFile, themePath))

	// Wait for change notification with timeout
	select {
	case <-changes:
		// Change detected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for change notification")
	}

	// Verify the palette was updated
	palette = cm.GetPalette()
	assert.Equal(t, "#NEWPRIMARY", palette.UI.BorderPrimary)

	// Do another atomic write + rename
	tmpFile2 := filepath.Join(tmpDir, "theme.json.tmp2")
	newThemeData2 := `{
		"themeName": "Test3",
		"isDark": true,
		"palette": {"primary": {"hex": "#ANOTHERPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(tmpFile2, []byte(newThemeData2), 0644))
	require.NoError(t, os.Rename(tmpFile2, themePath))

	// Wait for another change notification
	select {
	case <-changes:
		// Change detected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for second change notification")
	}

	// Verify the palette was updated again
	palette = cm.GetPalette()
	assert.Equal(t, "#ANOTHERPRIMARY", palette.UI.BorderPrimary)
}

func TestColorManager_LiveReload_CorruptFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create initial valid theme.json file
	themePath := filepath.Join(tmpDir, "theme.json")
	validThemeData := `{
		"themeName": "Test",
		"isDark": true,
		"palette": {"primary": {"hex": "#VALIDPRIMARY"}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`
	require.NoError(t, os.WriteFile(themePath, []byte(validThemeData), 0644))

	colorsPath := filepath.Join(tmpDir, "colors.json")

	// Create color manager
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Get initial palette
	palette := cm.GetPalette()
	assert.Equal(t, "#VALIDPRIMARY", palette.UI.BorderPrimary)

	// Set up channel to receive change notifications
	changes := cm.ColorChanges()

	// Modify the theme.json file to be corrupt
	tmpFile := filepath.Join(tmpDir, "theme.json.tmp")
	require.NoError(t, os.WriteFile(tmpFile, []byte("{invalid json"), 0644))
	require.NoError(t, os.Rename(tmpFile, themePath))

	// Wait for change notification
	select {
	case <-changes:
		// Change detected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for change notification on corrupt file")
	}

	// Should fall back to default
	palette = cm.GetPalette()
	assert.Equal(t, models.DefaultColorPalette().UI.BorderPrimary, palette.UI.BorderPrimary)
}

func TestColorManager_ThemeParentDirNotExists(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create theme path in non-existent directory
	themePath := filepath.Join(tmpDir, "nonexistent", "theme.json")
	colorsPath := filepath.Join(tmpDir, "colors.json")

	// This should not error
	cm, err := NewColorManagerWithPaths(themePath, colorsPath)
	require.NoError(t, err)
	defer cm.Close()

	// Should use default palette since theme path doesn't exist
	palette := cm.GetPalette()
	assert.Equal(t, models.DefaultColorPalette().UI.BorderPrimary, palette.UI.BorderPrimary)
}

func TestParseDTEThemeDocument_EmptyHex(t *testing.T) {
	// Test that empty hex values fall back to defaults
	jsonData := `{
		"themeName": "Test",
		"isDark": true,
		"palette": {"primary": {"hex": ""}},
		"dank16": {"colors": [], "background": {"hex": ""}, "foreground": {"hex": ""}, "cursor": {"hex": ""}},
		"generatedAt": 0
	}`

	doc, err := ParseDTEThemeDocument([]byte(jsonData))
	require.NoError(t, err)

	palette := doc.ToColorPalette()

	// Should fall back to default for empty hex
	assert.Equal(t, models.DefaultColorPalette().UI.BorderPrimary, palette.UI.BorderPrimary)
}

func TestDTEColor_UnmarshalJSON_BareString(t *testing.T) {
	var c DTEColor
	err := json.Unmarshal([]byte(`"#FF0000"`), &c)
	require.NoError(t, err)
	assert.Equal(t, "#FF0000", c.Hex)
	assert.Equal(t, "#FF0000", c.String())
}

func TestDTEColor_UnmarshalJSON_Object(t *testing.T) {
	var c DTEColor
	err := json.Unmarshal([]byte(`{"hex":"#00FF00","alpha":0.5}`), &c)
	require.NoError(t, err)
	assert.Equal(t, "#00FF00", c.Hex)
	assert.Equal(t, "#00FF00", c.String())
}
