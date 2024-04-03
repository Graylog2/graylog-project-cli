package utils

import "os"

// MavenBin returns the Maven command name.
// An existing Maven wrapper (./mvnw) in the current directory is preferred.
func MavenBin() string {
	// Prefer an existing Maven wrapper script to ensure consistent Maven version usage.
	mvnCmd := "./mvnw"
	if _, err := os.Stat(mvnCmd); os.IsNotExist(err) {
		mvnCmd = "mvn"
	}
	return mvnCmd
}
