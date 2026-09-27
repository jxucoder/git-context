package model

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/exec"
	"strings"
)

// GenerateID generates a random 8-character hex ID.
func GenerateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// GetAuthorShort returns the identity used for authorship, claims and locks:
// git's user.name, falling back to the OS user name, then "Unknown".
func GetAuthorShort() string {
	if name := GitConfig("user.name"); name != "" {
		return name
	}
	for _, key := range []string{"USER", "USERNAME"} {
		if name := strings.TrimSpace(os.Getenv(key)); name != "" {
			return name
		}
	}
	return "Unknown"
}

// GitConfig returns the value of a git config key, or "" when it is unset.
func GitConfig(key string) string {
	output, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
