package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
	name := repository

	scheme := regexp.MustCompile("^[a-zA-Z]+://")
	needle := scheme.FindStringIndex(name)
	if needle != nil {
		name = name[needle[1]:]
	}
	user := regexp.MustCompile("^[a-zA-Z0-9_.-:]+@")
	needle = user.FindStringIndex(name)
	if needle != nil {
		name = name[needle[1]:]
	}

	host := regexp.MustCompile("^[a-zA-Z0-9_.-]+[:/]")
	needle = host.FindStringIndex(name)
	if needle != nil {
		name = name[needle[1]:]
	}

	name = strings.Replace(path.Base(name), ".git", "", 1)

	if name == "." || name == "/" || name == "" {
		 logger.Fatal("Unable to get name from repository: %s", repository)
	}

	return name

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

func SetPackageJsonVersion(filename, version string) error {
	// Make sure to match the correct version string in package.json
	re := regexp.MustCompile(`^\s{2}"version": "\d+\.\d+\.\d+-?.*?",`)

	err := ReplaceInFile(filename, re, fmt.Sprintf(`  "version": "%s",`, version))
	if err != nil {
		return err
	}

	return nil
}

func ReplaceInFile(filename string, re *regexp.Regexp, replacement string) error {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(absFilename)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadFile(absFilename)
	if err != nil {
		return err
	}

	lines := make([]string, 0)

	for _, line := range strings.Split(string(buf), "\n") {
		if re.MatchString(line) {
			lines = append(lines, re.ReplaceAllString(line, replacement))
		} else {
			lines = append(lines, line)
		}
	}

	f, err := ioutil.TempFile(filepath.Dir(absFilename), fileInfo.Name())
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(strings.Join(lines, "\n")))
	if err != nil {
		return nil
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(f.Name(), absFilename); err != nil {
		return err
	}

	if err := os.Chmod(absFilename, fileInfo.Mode()); err != nil {
		return err
	}

	return nil
}
