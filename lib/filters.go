package lib 

import (
	"strings"
	"unicode"
)

// Pass a string through a series of filters
func Filter(str string, filterFuncs []string) string {
	for _, name := range filterFuncs {
		if name == "remove-spaces" {
			str = RemoveSpaces(str)
			continue
		}
	}
	return str
} 

// Removes all white spaces in the string
func RemoveSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}