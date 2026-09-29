package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	illegalChars  = regexp.MustCompile(`[<>:"/\\|?*]`)
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

// ValidateName checks a display name, which also becomes the name of its files.
func ValidateName(name string) error {
	switch {
	case name == "" || utf8.RuneCountInString(name) > 255:
		return errors.New("name must be between 1 and 255 characters")
	case illegalChars.MatchString(name) || controlChars.MatchString(name):
		return errors.New(`name cannot contain < > : " / \ | ? * or control characters`)
	case strings.Trim(name, ".") == "" || reservedNames.MatchString(name):
		return errors.New("name is reserved by the file system")
	}
	return nil
}
