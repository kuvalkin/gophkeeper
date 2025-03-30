package utils

import (
	"crypto/sha256"
	"fmt"
)

func GetEntryKey(prefix string, name string) string {
	hash := sha256.Sum256([]byte(prefix + name))

	return fmt.Sprintf("%x", hash)
}
