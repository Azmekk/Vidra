package utils

import (
	"regexp"
	"strings"
)

var (
	illegalChars = regexp.MustCompile(`[<>:"/\\|?*]`) // Corrected escaping for backslash
	controlChars = regexp.MustCompile(`[\x00-\x1f\x7f]`) // Corrected escaping for backslash
	reservedNames = regexp.MustCompile(`^(?i)(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\..*)?$`) // Corrected escaping for backslash
)

// SanitizeFilename removes or replaces characters that are illegal in filenames across most OSs
func SanitizeFilename(name string) string {
	// Replace illegal characters with underscores
	name = illegalChars.ReplaceAllString(name, "_")
	// Replace control characters
	name = controlChars.ReplaceAllString(name, "_")
	
	// Trim spaces and dots (problematic on Windows at end of name)
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	
	// If empty or reserved name, use a default
	if name == "" || reservedNames.MatchString(name) {
		return "video"
	}

	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}

	return name
}
