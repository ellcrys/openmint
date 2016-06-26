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

// Pass a slice of strings to a series of slice filters
func MatchFilter(matches []string, filterFuncs []string) []string {
	for _, name := range filterFuncs {
		if name == "remove-empty" {
			matches = SliceRemoveEmpty(matches)
		}
	}
	return matches
}

// Removes all whitespace elements from an array of strings
func SliceRemoveEmpty(strs []string) []string {
	var newStrs []string
	for _, s := range strs {
		if s != "" {
			newStrs = append(newStrs, s)
		}
	}
	return newStrs
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