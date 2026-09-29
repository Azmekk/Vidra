package utils

import (
	"regexp"
	"strings"
)

var (
	illegalChars  = regexp.MustCompile(`[<>:"/\|?*]`)
	controlChars  = regexp.MustCompile(`[\x00-\x1f\x7f]`)
	reservedNames = regexp.MustCompile(`^(?i)(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\..*)?$`)
)

// SanitizeFilename replaces characters that are illegal in filenames across common OSs.
func SanitizeFilename(name string) string {
	name = illegalChars.ReplaceAllString(name, "_")
	name = controlChars.ReplaceAllString(name, "_")
	name = strings.Trim(strings.TrimSpace(name), ".")

	if name == "" || reservedNames.MatchString(name) {
		return "video"
	}
	if len(name) > 200 {
		name = strings.ToValidUTF8(name[:200], "")
	}
	return name
}
