package tui

import (
	"runtime"
	"strings"

	"github.com/JonasNowak/dms-gop/models"
)

type distroArt struct {
	matches []string
	color   string
	art     []string
}

var genericArt = distroArt{
	color: "#7D56F4",
	art: []string{
		"    ___",
		"   (.. \\\"",
		"   (<> |",
		"  //  \\\\ \\\\",
		" ( |  | /|",
		"_/\\\\ __)/_)",
		"\\\\/____\\\\/",
	},
}

// appleArt is the macOS logo. Its color is left empty so the renderer falls
// back to the theme accent color instead of a hardcoded white.
var appleArt = distroArt{
	matches: []string{"macos", "mac os", "darwin"},
	art: []string{
		"       .:'",
		"    __ :'__",
		" .'`  `-'  ``.",
		":          .-'",
		":         :",
		" :         `-;",
		"  `.__.-.__.'",
	},
}

func getDistroInfo(hardware *models.SystemHardware) ([]string, string) {
	if runtime.GOOS == "darwin" {
		return appleArt.art, appleArt.color
	}
	if hardware != nil {
		distro := strings.ToLower(hardware.Distro)
		for _, match := range appleArt.matches {
			if strings.Contains(distro, match) {
				return appleArt.art, appleArt.color
			}
		}
	}
	return genericArt.art, genericArt.color
}
