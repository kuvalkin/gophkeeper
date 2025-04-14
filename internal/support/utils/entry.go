// Package utils provides utility functions and helpers for common operations.
package utils

import (
	"crypto/sha256"
	"fmt"
)

// GetEntryKey generates a SHA-256 hash of the concatenation of the prefix and name,
// and returns it as a hexadecimal string.
func GetEntryKey(prefix string, name string) string {
	hash := sha256.Sum256([]byte(prefix + name))

	return fmt.Sprintf("%x", hash)
}
