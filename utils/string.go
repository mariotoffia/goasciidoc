package utils

import "strings"

// ContainsString checks if _s_ is in the array _slz_.
//
// If fails, it returns -1 and `false`.
func ContainsString(slz []string, s string) (int, bool) {

	for i := range slz {

		if slz[i] == s {
			return i, true
		}

	}

	return -1, false
}

// HasSuffixString returns the first string in _slz_ that has
// suffix of _s_.
//
// If fails, it returns -1 and `false`.
func HasSuffixString(slz []string, s string) (int, bool) {

	for i := range slz {

		if strings.HasSuffix(slz[i], s) {
			return i, true
		}

	}

	return -1, false
}

// RemoveString will remove a string from _s_ at _index_.
func RemoveString(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
