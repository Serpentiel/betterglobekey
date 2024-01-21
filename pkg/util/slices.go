// Package util contains utility functions.
package util

// Contains checks if a slice contains a given string.
func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}

	return false
}
