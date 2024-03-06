package git

import (
	"fmt"
	"os"
)

func GetRemoteUrl(path string, remote string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(path); err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	urlString, err := GitValueE("remote", "get-url", "--push", remote)
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	return urlString, nil
}
