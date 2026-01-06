package model

import (
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"strings"
)

// GenerateID generates a random 8-character hex ID.
func GenerateID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetAuthor returns the git user name and email as "Name <email>".
func GetAuthor() string {
	name := getGitConfig("user.name")
	email := getGitConfig("user.email")
	
	if name == "" {
		name = "Unknown"
	}
	if email != "" {
		return name + " <" + email + ">"
	}
	return name
}

// GetAuthorShort returns just the git user name.
func GetAuthorShort() string {
	name := getGitConfig("user.name")
	if name == "" {
		return "Unknown"
	}
	return name
}

func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

