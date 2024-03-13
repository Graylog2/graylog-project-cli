package selfupdate

import (
	"context"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/ask"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/hashicorp/go-version"
	"github.com/mattn/go-isatty"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner = "Graylog2"
	repoName  = "graylog-project-cli"
)

func SelfUpdate(runningVersion *version.Version, requestedVersion string, force bool, interactive bool) error {
	// Find the current binary first
	binPath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return fmt.Errorf("couldn't find binary path for %q: %w", os.Args[0], err)
	}

	binFileInfo, err := os.Lstat(binPath)
	if err != nil {
		return fmt.Errorf("couldn't get file info for current binary %q: %w", binPath, err)
	}

	if binFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(binPath)
		if err != nil {
			return fmt.Errorf("couldn't resolve symlink for current binary %q: %w", binPath, err)
		}

		binPath, err = filepath.Abs(linkTarget)
		if err != nil {
			return fmt.Errorf("couldn't resolve absolute path for symlink target %q: %w", linkTarget, err)
		}
	}

	client := github.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Getting latest release info...")

	var release *github.RepositoryRelease
	var response *github.Response

	if requestedVersion != "latest" {
		release, response, err = client.Repositories.GetReleaseByTag(ctx, repoOwner, repoName, requestedVersion)
	} else {
		release, response, err = client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	}
	if err != nil {
		if response.StatusCode == 404 {
			return fmt.Errorf("version %s doesn't exist", requestedVersion)
		}
		return fmt.Errorf("couldn't get release: %w", err)
	}

	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("couldn't read error body for status %q: %w", response.Status, err)
		}
		return fmt.Errorf("couldn't list releases: %s\n%s", response.Status, string(body))
	}

	latestVersion, err := version.NewVersion(release.GetTagName())
	if err != nil {
		return fmt.Errorf("couldn't parse latest version %q: %w", release.GetTagName(), err)
	}

	if latestVersion.LessThanOrEqual(runningVersion) && !force {
		logger.ColorInfo(color.FgGreen, "You are running the latest version: %s", runningVersion)
		return nil
	}

	logger.Info("Updating: %s", binPath)

	if interactive && isatty.IsTerminal(os.Stdout.Fd()) {
		asker := ask.NewAsker(os.Stdin)
		if !asker.AskYesNo(color.YellowString("Changelog: %s\nUpdate to version %s?", *release.HTMLURL, latestVersion), true) {
			return nil
		}
	}

	logger.ColorInfo(color.FgGreen, "Updating to %s", latestVersion)

	osAssets := lo.Filter(release.Assets, func(item *github.ReleaseAsset, index int) bool {
		if item.Name == nil {
			panic(fmt.Errorf("GitHub asset name cannot be nil"))
		}
		// We only have amd64 binaries for linux as of 2024-02-05
		//goland:noinspection GoBoolExpressions
		return strings.Contains(*item.Name, runtime.GOOS) && (strings.Contains(*item.Name, runtime.GOARCH) || runtime.GOOS == "linux")
	})

	if len(osAssets) > 1 {
		names := strings.Join(lo.Map(osAssets, func(item *github.ReleaseAsset, index int) string {
			return item.GetName()
		}), ", ")
		return fmt.Errorf("found more than one asset for %s/%s: %v", runtime.GOOS, runtime.GOARCH, names)
	} else if len(osAssets) < 1 {
		return fmt.Errorf("found no asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	osAsset := osAssets[0]
	downloadUrl := osAsset.GetBrowserDownloadURL()

	if strings.TrimSpace(downloadUrl) == "" {
		return fmt.Errorf("release download URL cannot be empty")
	}

	newFile, err := os.CreateTemp(filepath.Dir(binPath), osAsset.GetName()+"-*")
	if err != nil {
		return fmt.Errorf("couldn't create temporary file: %w", err)
	}

	if err := newFile.Chmod(binFileInfo.Mode()); err != nil {
		_ = newFile.Close()
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't set file mode on %q: %w", newFile.Name(), err)
	}

	downloadResponse, err := (&http.Client{Timeout: 5 * time.Minute}).Get(downloadUrl)
	if err != nil {
		_ = newFile.Close()
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't download from URL %q, %w", downloadUrl, err)
	}
	defer downloadResponse.Body.Close()

	bar := progressbar.DefaultBytes(downloadResponse.ContentLength, "Downloading "+osAsset.GetName())

	if _, err := io.Copy(io.MultiWriter(newFile, bar), downloadResponse.Body); err != nil {
		_ = newFile.Close()
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't write download response: %w", err)
	}
	if err := newFile.Sync(); err != nil {
		_ = newFile.Close()
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't sync downloaded file %q: %w", newFile.Name(), err)
	}
	if err := newFile.Close(); err != nil {
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't close downloaded file %q: %w", newFile.Name(), err)
	}

	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		// Running binaries on Windows are locked and can't be removed/overwritten while still running.
		// We are renaming the running binary to avoid errors.

		bakFile := filepath.Join(filepath.Dir(binPath), filepath.Base(binPath)+".bak")
		if _, err := os.Stat(bakFile); err == nil {
			if err := os.Remove(bakFile); err != nil {
				return fmt.Errorf("couldn't remove existing .bak file %q: %w", bakFile, err)
			}
		}
		if err := os.Rename(binPath, bakFile); err != nil {
			return fmt.Errorf("couldn't rename %q: %w", binPath, err)
		}
	}

	if err := os.Rename(newFile.Name(), binPath); err != nil {
		_ = os.Remove(newFile.Name())
		return fmt.Errorf("couldn't rename downloaded file to %q: %w", binPath, err)
	}

	logger.ColorInfo(color.FgGreen, "Done - Release notes: %s", *release.HTMLURL)

	return nil
}
