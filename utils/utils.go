package utils

import (
	"bytes"
	"errors"
	"github.com/Graylog2/graylog-project-cli/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func GetCwd() string {
	currentDir, err := os.Getwd()

	if err != nil {
		logger.Fatal("Unable to get current directory: %v", err)
	}

	// Make sure to resolve any symlinks and get the real directory.
	// Functions like filepath.Walk do not handle symlinks well...
	dir, err := filepath.EvalSymlinks(currentDir)
	if err != nil {
		logger.Fatal("Unable to eval symlink for %v: %v", currentDir, err)
	}

	return dir
}

func Chdir(path string) {
	if err := os.Chdir(path); err != nil {
		logger.Fatal("Unable to change into directory %v: %v", path, err)
	}
}

// Returns the relative path from the current working directory for the given path.
func GetRelativePath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}

	cwd := GetCwd()
	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		logger.Fatal("Unable to get relative path for %v", path)
	}

	return relPath
}

// Returns the relative path from the current working directory for the given path.
// It also evaluates symlinks in the given path before it returns the relative path.
func GetRelativePathEvalSymlinks(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}

	newPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		logger.Fatal("Unable to eval symlinks for %v: %v", path, err)
	}

	return GetRelativePath(newPath)
}

func GetAbsolutePath(path string) string {
	absolutePath, err := filepath.Abs(path)

	if err != nil {
		logger.Fatal("Unable to get absolute path for %s: %v", path, err)
	}

	return absolutePath
}

func NameFromRepository(repository string) string {
	if strings.HasPrefix(repository, "https://") {
		return strings.Replace(strings.Split(strings.TrimPrefix(repository, "https://"), "/")[2], ".git", "", 1)
	} else if strings.HasPrefix(repository, "git@") {
		return strings.Replace(strings.Split(repository, "/")[1], ".git", "", 1)
	} else {
		logger.Fatal("Unable to get name from repository: %s", repository)
	}
	return ""
}

func ConvertGithubGitToHTTPS(repository string) string {
	return strings.Replace(repository, "git@github.com:", "https://github.com/", 1)
}

func FirstNonEmpty(values ...string) (string, error) {
	for idx := range values {
		trimmedValue := strings.TrimSpace(values[idx])

		if trimmedValue != "" {
			return values[idx], nil
		}
	}

	return "", errors.New("all values are empty")
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

type DirectoryCallback func()

func InDirectory(path string, callback DirectoryCallback) {
	defer Chdir(GetCwd())

	Chdir(path)

	callback()
}

func FilesIdentical(file1, file2 string) bool {
	buf1, err := ioutil.ReadFile(file1)
	if err != nil {
		logger.Fatal("Unable to read file <%s>: %v", file1, err)
	}
	buf2, err := ioutil.ReadFile(file2)
	if err != nil {
		logger.Fatal("Unable to read file <%s>: %v", file2, err)
	}

	return bytes.Equal(buf1, buf2)
}
