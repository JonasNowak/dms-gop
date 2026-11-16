package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JonasNowak/dms-gop/models"
)

// DTEColor represents a color from the Dank Theme Engine that can be either
// a string hex value or an object with hex and other properties.
type DTEColor struct {
	Hex       string `json:"hex"`
	RawString string // holds the raw string value when unmarshaling bare hex
}

// UnmarshalJSON implements json.Unmarshaler for DTEColor.
// It accepts either a bare hex string like "#RRGGBB" or an object like {"hex":"#RRGGBB",...}.
func (c *DTEColor) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err == nil {
		c.RawString = hexStr
		c.Hex = hexStr
		return nil
	}

	// Try to unmarshal as an object
	var obj struct {
		Hex string `json:"hex"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		c.Hex = obj.Hex
		return nil
	}

	return fmt.Errorf("invalid DTEColor format: %s", string(data))
}

// MarshalJSON implements json.Marshaler for DTEColor.
func (c DTEColor) MarshalJSON() ([]byte, error) {
	if c.RawString != "" {
		return json.Marshal(c.RawString)
	}
	return json.Marshal(map[string]string{"hex": c.Hex})
}

// String returns the hex value of the color.
func (c *DTEColor) String() string {
	if c.Hex != "" {
		return c.Hex
	}
	return c.RawString
}

// DTEPalette represents the Material You 3 color palette from dank-theme-engine.
type DTEPalette struct {
	Name                    string   `json:"name"`
	IsDark                  bool     `json:"isDark"`
	Primary                 DTEColor `json:"primary"`
	OnPrimary               DTEColor `json:"onPrimary"`
	PrimaryContainer        DTEColor `json:"primaryContainer"`
	OnPrimaryContainer      DTEColor `json:"onPrimaryContainer"`
	InversePrimary          DTEColor `json:"inversePrimary"`
	Secondary               DTEColor `json:"secondary"`
	OnSecondary             DTEColor `json:"onSecondary"`
	SecondaryContainer      DTEColor `json:"secondaryContainer"`
	OnSecondaryContainer    DTEColor `json:"onSecondaryContainer"`
	Tertiary                DTEColor `json:"tertiary"`
	OnTertiary              DTEColor `json:"onTertiary"`
	TertiaryContainer       DTEColor `json:"tertiaryContainer"`
	OnTertiaryContainer     DTEColor `json:"onTertiaryContainer"`
	Surface                 DTEColor `json:"surface"`
	OnSurface               DTEColor `json:"onSurface"`
	SurfaceVariant          DTEColor `json:"surfaceVariant"`
	OnSurfaceVariant        DTEColor `json:"onSurfaceVariant"`
	SurfaceTint             DTEColor `json:"surfaceTint"`
	SurfaceContainerLowest  DTEColor `json:"surfaceContainerLowest"`
	SurfaceContainerLow     DTEColor `json:"surfaceContainerLow"`
	SurfaceContainer        DTEColor `json:"surfaceContainer"`
	SurfaceContainerHigh    DTEColor `json:"surfaceContainerHigh"`
	SurfaceContainerHighest DTEColor `json:"surfaceContainerHighest"`
	Background              DTEColor `json:"background"`
	OnBackground            DTEColor `json:"onBackground"`
	Outline                 DTEColor `json:"outline"`
	OutlineVariant          DTEColor `json:"outlineVariant"`
	Error                   DTEColor `json:"error"`
	OnError                 DTEColor `json:"onError"`
	ErrorContainer          DTEColor `json:"errorContainer"`
	OnErrorContainer        DTEColor `json:"onErrorContainer"`
	Warning                 DTEColor `json:"warning"`
	OnWarning               DTEColor `json:"onWarning"`
	WarningContainer        DTEColor `json:"warningContainer"`
	OnWarningContainer      DTEColor `json:"onWarningContainer"`
	Info                    DTEColor `json:"info"`
	OnInfo                  DTEColor `json:"onInfo"`
	InfoContainer           DTEColor `json:"infoContainer"`
	OnInfoContainer         DTEColor `json:"onInfoContainer"`
	Success                 DTEColor `json:"success"`
	OnSuccess               DTEColor `json:"onSuccess"`
	SuccessContainer        DTEColor `json:"successContainer"`
	OnSuccessContainer      DTEColor `json:"onSuccessContainer"`
	Scrim                   DTEColor `json:"scrim"`
	Shadow                  DTEColor `json:"shadow"`
}

// DTE16 represents the 16-color ansi palette from dank-theme-engine.
type DTE16 struct {
	Colors     []DTEColor `json:"colors"`
	Background DTEColor   `json:"background"`
	Foreground DTEColor   `json:"foreground"`
	Cursor     DTEColor   `json:"cursor"`
}

// DTEThemeDocument represents the complete theme document from dank-theme-engine.
type DTEThemeDocument struct {
	ThemeName      string     `json:"themeName"`
	IsDark         bool       `json:"isDark"`
	Palette        DTEPalette `json:"palette"`
	Dank16         DTE16      `json:"dank16"`
	GeneratedAt    float64    `json:"generatedAt"` // Can be float epoch seconds OR we handle string separately
	GeneratedAtStr string     `json:"-"`           // For internal use when parsing RFC3339
}

// UnmarshalJSON implements custom unmarshaling for DTEThemeDocument to handle
// both float epoch seconds and RFC3339 string for generatedAt.
func (d *DTEThemeDocument) UnmarshalJSON(data []byte) error {
	// First, parse into a temporary struct that has GeneratedAt as json.RawMessage
	type tempDoc struct {
		ThemeName   string          `json:"themeName"`
		IsDark      bool            `json:"isDark"`
		Palette     DTEPalette      `json:"palette"`
		Dank16      DTE16           `json:"dank16"`
		GeneratedAt json.RawMessage `json:"generatedAt"`
	}
	var temp tempDoc
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	d.ThemeName = temp.ThemeName
	d.IsDark = temp.IsDark
	d.Palette = temp.Palette
	d.Dank16 = temp.Dank16

	// Try to parse generatedAt as float
	var genAtFloat float64
	if err := json.Unmarshal(temp.GeneratedAt, &genAtFloat); err == nil {
		d.GeneratedAt = genAtFloat
		return nil
	}

	// Try to parse as string (RFC3339)
	var genAtStr string
	if err := json.Unmarshal(temp.GeneratedAt, &genAtStr); err == nil {
		d.GeneratedAtStr = genAtStr
		// Try to parse the RFC3339 string
		if t, err := time.Parse(time.RFC3339, genAtStr); err == nil {
			d.GeneratedAt = float64(t.Unix())
		} else if t, err := time.Parse(time.RFC3339Nano, genAtStr); err == nil {
			d.GeneratedAt = float64(t.Unix())
		}
		return nil
	}

	// If neither works, that's fine - it's optional
	return nil
}

// getColor returns the hex string from a DTEColor, falling back to default if empty.
func getColor(c DTEColor, defaultColor string) string {
	hex := c.String()
	if hex == "" || hex == "#" || len(hex) < 4 {
		return defaultColor
	}
	// Ensure it's a valid hex color (starts with # and has 3, 4, 6, or 8 chars)
	if !strings.HasPrefix(hex, "#") {
		return defaultColor
	}
	return hex
}

// getDank16Color returns the hex string from dank16 colors array at index i,
// falling back to default if the index is out of range or the color is invalid.
func getDank16Color(d *DTE16, index int, defaultColor string) string {
	if d == nil || index < 0 || index >= len(d.Colors) {
		return defaultColor
	}
	return getColor(d.Colors[index], defaultColor)
}

// ParseDTEThemeDocument parses the theme.json file produced by dank-theme-engine / DankMacShell.
func ParseDTEThemeDocument(data []byte) (*DTEThemeDocument, error) {
	var doc DTEThemeDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse DTE theme document: %w", err)
	}

	return &doc, nil
}

// ToColorPalette converts a DTEThemeDocument to a models.ColorPalette,
// mapping M3 + dank16 colors onto every slot with fallback to defaults.
func (d *DTEThemeDocument) ToColorPalette() *models.ColorPalette {
	palette := models.DefaultColorPalette()

	// Helper to get color with fallback
	fallback := func(c DTEColor, defaultColor string) string {
		return getColor(c, defaultColor)
	}

	// UI colors
	palette.UI.BorderPrimary = fallback(d.Palette.Primary, palette.UI.BorderPrimary)
	palette.UI.BorderSecondary = fallback(d.Palette.InversePrimary, palette.UI.BorderSecondary)
	palette.UI.HeaderBackground = fallback(d.Palette.PrimaryContainer, palette.UI.HeaderBackground)
	palette.UI.HeaderText = fallback(d.Palette.OnPrimaryContainer, palette.UI.HeaderText)
	palette.UI.FooterBackground = fallback(d.Palette.SurfaceContainer, palette.UI.FooterBackground)
	palette.UI.FooterText = fallback(d.Palette.OnSurfaceVariant, palette.UI.FooterText)
	palette.UI.TextPrimary = fallback(d.Palette.OnSurface, palette.UI.TextPrimary)
	palette.UI.TextSecondary = fallback(d.Palette.OnSurfaceVariant, palette.UI.TextSecondary)
	palette.UI.TextAccent = fallback(d.Palette.Primary, palette.UI.TextAccent)
	palette.UI.SelectionBackground = fallback(d.Palette.SurfaceContainerHigh, palette.UI.SelectionBackground)
	palette.UI.SelectionText = fallback(d.Palette.OnSurface, palette.UI.SelectionText)

	// Chart colors
	// NetworkDownload = dank16.Colors[6] (or info)
	palette.Charts.NetworkDownload = getDank16Color(&d.Dank16, 6, fallback(d.Palette.Info, palette.Charts.NetworkDownload))
	// NetworkUpload = dank16.Colors[5] (or tertiary)
	palette.Charts.NetworkUpload = getDank16Color(&d.Dank16, 5, fallback(d.Palette.Tertiary, palette.Charts.NetworkUpload))
	// NetworkLine = outline
	palette.Charts.NetworkLine = fallback(d.Palette.Outline, palette.Charts.NetworkLine)
	// CPUCoreLow = primary
	palette.Charts.CPUCoreLow = fallback(d.Palette.Primary, palette.Charts.CPUCoreLow)
	// CPUCoreMedium = secondary
	palette.Charts.CPUCoreMedium = fallback(d.Palette.Secondary, palette.Charts.CPUCoreMedium)
	// CPUCoreHigh = tertiary
	palette.Charts.CPUCoreHigh = fallback(d.Palette.Tertiary, palette.Charts.CPUCoreHigh)
	// DiskRead = info
	palette.Charts.DiskRead = fallback(d.Palette.Info, palette.Charts.DiskRead)
	// DiskWrite = secondary
	palette.Charts.DiskWrite = fallback(d.Palette.Secondary, palette.Charts.DiskWrite)

	// ProgressBars
	// CPU, Memory, Disk: Low=primary, Medium=tertiary, High=error
	palette.ProgressBars.CPULow = fallback(d.Palette.Primary, palette.ProgressBars.CPULow)
	palette.ProgressBars.CPUMedium = fallback(d.Palette.Tertiary, palette.ProgressBars.CPUMedium)
	palette.ProgressBars.CPUHigh = fallback(d.Palette.Error, palette.ProgressBars.CPUHigh)

	palette.ProgressBars.MemoryLow = fallback(d.Palette.Primary, palette.ProgressBars.MemoryLow)
	palette.ProgressBars.MemoryMedium = fallback(d.Palette.Tertiary, palette.ProgressBars.MemoryMedium)
	palette.ProgressBars.MemoryHigh = fallback(d.Palette.Error, palette.ProgressBars.MemoryHigh)

	palette.ProgressBars.DiskLow = fallback(d.Palette.Primary, palette.ProgressBars.DiskLow)
	palette.ProgressBars.DiskMedium = fallback(d.Palette.Tertiary, palette.ProgressBars.DiskMedium)
	palette.ProgressBars.DiskHigh = fallback(d.Palette.Error, palette.ProgressBars.DiskHigh)

	// ProgressBackground = surfaceContainerHighest
	palette.ProgressBars.ProgressBackground = fallback(d.Palette.SurfaceContainerHighest, palette.ProgressBars.ProgressBackground)

	// Temperature
	// Cold = dank16.Colors[4] (with info fallback)
	palette.Temperature.Cold = getDank16Color(&d.Dank16, 4, fallback(d.Palette.Info, palette.Temperature.Cold))
	// Warm = dank16.Colors[3] (with warning fallback)
	palette.Temperature.Warm = getDank16Color(&d.Dank16, 3, fallback(d.Palette.Warning, palette.Temperature.Warm))
	// Hot = dank16.Colors[1] (with error fallback)
	palette.Temperature.Hot = getDank16Color(&d.Dank16, 1, fallback(d.Palette.Error, palette.Temperature.Hot))
	// Danger = dank16.Colors[9] (with error fallback)
	palette.Temperature.Danger = getDank16Color(&d.Dank16, 9, fallback(d.Palette.Error, palette.Temperature.Danger))

	// Status colors from palette
	palette.Status.Success = fallback(d.Palette.Success, palette.Status.Success)
	palette.Status.Warning = fallback(d.Palette.Warning, palette.Status.Warning)
	palette.Status.Error = fallback(d.Palette.Error, palette.Status.Error)
	palette.Status.Info = fallback(d.Palette.Info, palette.Status.Info)

	return palette
}
