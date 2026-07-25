package main

import (
	"strings"
)

func matchesAnyPattern(filename string, patterns []string) bool {
	lower := strings.ToLower(filename)
	for _, p := range patterns {
		if matchGlob(strings.ToLower(p), lower) {
			return true
		}
	}
	return false
}

func matchGlob(pattern, name string) bool {
	for {
		if pattern == "" {
			return name == ""
		}
		if pattern == "*" {
			return true
		}

		i := strings.Index(pattern, "*")
		if i < 0 {
			return strings.HasSuffix(name, pattern)
		}

		prefix := pattern[:i]
		if !strings.HasPrefix(name, prefix) {
			return false
		}

		name = name[len(prefix):]
		pattern = pattern[i+1:]
	}
}
