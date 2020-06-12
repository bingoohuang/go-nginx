package util

import "strings"

// TryAppend tries to append the given suffix if it does not ends in the s.
func TryAppend(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}

	return s + suffix
}

// TryPrepend tries to prepend the given prefix if it does not start in the s.
func TryPrepend(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}

	return prefix + s
}
