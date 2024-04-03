package utils

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	var url GitHubURL
	var expected string
	var err error

	// Custom URL
	url, _ = ParseGitHubURL("github://Graylog2/graylog2-server.git")
	expected = "git@github.com:Graylog2/graylog2-server.git"
	if url.SSH() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.SSH())
	}

	url, _ = ParseGitHubURL("github://Graylog2/graylog2-server.git")
	expected = "https://github.com/Graylog2/graylog2-server.git"
	if url.HTTPS() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.HTTPS())
	}

	// Git URL
	url, _ = ParseGitHubURL("git@github.com:Graylog2/graylog2-server.git")
	expected = "git@github.com:Graylog2/graylog2-server.git"
	if url.SSH() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.SSH())
	}

	url, _ = ParseGitHubURL("git@github.com:Graylog2/graylog2-server.git")
	expected = "https://github.com/Graylog2/graylog2-server.git"
	if url.HTTPS() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.HTTPS())
	}

	// HTTPS URL
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	expected = "git@github.com:Graylog2/graylog2-server.git"
	if url.SSH() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.SSH())
	}

	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	expected = "https://github.com/Graylog2/graylog2-server.git"
	if url.HTTPS() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.HTTPS())
	}

	// Directory
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	expected = "graylog2-server"
	if url.Directory() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.HTTPS())
	}

	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	expected = "https://github.com/Graylog2/graylog2-server"
	if url.BrowserURL() != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url.BrowserURL())
	}

	// Matches ok
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	match := "Graylog2/graylog2-server"
	if !url.Matches(match) {
		t.Errorf("expected <%s> to match <%s>", url, match)
	}

	// Matches case
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	match = "GraYloG2/GraYlog2-SErver"
	if !url.Matches(match) {
		t.Errorf("expected <%s> to match <%s>", url, match)
	}

	// Match fails
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	match = "Graylog2/graylog2-server-does-not-work"
	if url.Matches(match) {
		t.Errorf("expected <%s> to not match <%s>", url, match)
	}

	// Missing .git suffix
	_, err = ParseGitHubURL("https://github.com/Graylog2/graylog2-server")
	if err == nil {
		t.Error("expected URL without .git suffix to fail")
	}

	// With authentication
	url, err = ParseGitHubURL("https://user:password@github.com/Graylog2/graylog2-server.git")
	match = "https://github.com/Graylog2/graylog2-server.git"
	if err != nil {
		t.Errorf("expected URL with user:password not to fail: %s", err)
	}
	if url.URL() != match {
		t.Errorf("expected <%s> to be <%s>", url.URL(), match)
	}

	// Unknown URL format
	_, err = ParseGitHubURL("https://example.com/Graylog2/graylog2-server")
	if err == nil {
		t.Error("expected unknown URL to fail")
	}

	url, _ = ParseGitHubURL("github://Graylog2/graylog2-server.git")
	if url.IsSSH() || url.IsHTTPS() {
		t.Error("expected URL to not be SSH or HTTPS")
	}
	url, _ = ParseGitHubURL("https://github.com/Graylog2/graylog2-server.git")
	if !url.IsHTTPS() || url.IsSSH() {
		t.Error("expected URL to be HTTPS and not SSH")
	}
	url, _ = ParseGitHubURL("git@github.com:Graylog2/graylog2-server.git")
	if url.IsHTTPS() || !url.IsSSH() {
		t.Error("expected URL to be SSH and not HTTPS")
	}
}

func TestReplaceGitHubURL(t *testing.T) {
	url, _ := ReplaceGitHubURL("github://Graylog2/graylog2-server.git", "foo/graylog2-server")
	expected := "github://foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}
	url, _ = ReplaceGitHubURL("github://Graylog2/graylog2-server.git", "foo/graylog2-server.git")
	expected = "github://foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}

	url, _ = ReplaceGitHubURL("https://github.com/Graylog2/graylog2-server.git", "foo/graylog2-server")
	expected = "https://github.com/foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}
	url, _ = ReplaceGitHubURL("https://github.com/Graylog2/graylog2-server.git", "foo/graylog2-server.git")
	expected = "https://github.com/foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}

	url, _ = ReplaceGitHubURL("https://github.com/Graylog2/graylog2-server.git", "foo/graylog2-server")
	expected = "https://github.com/foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}
	url, _ = ReplaceGitHubURL("https://github.com/Graylog2/graylog2-server.git", "foo/graylog2-server.git")
	expected = "https://github.com/foo/graylog2-server.git"
	if url != expected {
		t.Errorf("expected <%s> but got <%s>", expected, url)
	}
}

func TestParseGitHubPRString(t *testing.T) {
	var cases = []struct {
		input    string
		prRepo   string
		prNumber int
		err      bool
	}{
		{"Graylog2/graylog2-server#123", "Graylog2/graylog2-server", 123, false},
		{"https://github.com/Graylog2/graylog-plugin-collector/pull/9692", "Graylog2/graylog-plugin-collector", 9692, false},
		{"https://github.com/Graylog2/graylog-plugin-collector/pull/", "", 0, true},
		{"https://github.com/9692", "", 0, true},
		{"https://example.com/Graylog2/graylog-plugin-collector/pull/9692", "", 0, true},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			repo, num, err := ParseGitHubPRString(c.input)
			if repo != c.prRepo {
				t.Errorf("expected <%s>, got <%s>", c.prRepo, repo)
			}
			if num != c.prNumber {
				t.Errorf("expected <%d>, got <%d>", c.prNumber, num)
			}
			if err == nil && c.err {
				t.Errorf("expected an error but got none")
			}
			if err != nil && !c.err {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestResolveGitHubIssueURL(t *testing.T) {
	var cases = []struct {
		baseRepo string
		input    string
		output   string
		err      bool
	}{
		{"https://github.com/myorg/repo1.git", "123", "https://github.com/myorg/repo1/issues/123", false},
		{"https://github.com/myorg/repo1.git", "#123", "https://github.com/myorg/repo1/issues/123", false},
		{"https://github.com/myorg/repo1.git", "repo1#123", "https://github.com/myorg/repo1/issues/123", false},
		{"https://github.com/myorg/repo1.git", "repo2#123", "https://github.com/myorg/repo2/issues/123", false},
		{"https://github.com/myorg/repo1.git", "myorg/repo1#123", "https://github.com/myorg/repo1/issues/123", false},
		{"https://github.com/myorg/repo1.git", "myorg/repo2#123", "https://github.com/myorg/repo2/issues/123", false},
		{"https://github.com/myorg/repo1.git", "https://github.com/myorg/repo1/issues/123", "https://github.com/myorg/repo1/issues/123", false},
		{"https://github.com/myorg/repo1.git", "https://github.com/myorg/repo1/pull/123", "https://github.com/myorg/repo1/pull/123", false},
		{"https://github.com/myorg/repo1.git", "abc123", "", true},
		{"https://github.com/myorg/repo1.git", "abc123", "", true},
		{"https://github.com/myorg/repo1.git", "https://github.com/myorg/repo1", "", true},
		{"https://github.com/myorg/repo1.git", "https://github.com/myorg/repo1/yolo/123", "", true},
		{"https://github.com/myorg/repo1.git", "", "", true},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			url, err := ResolveGitHubIssueURL(c.baseRepo, c.input)
			if url != c.output {
				t.Errorf("expected <%s> for input <%s>, got <%s>", c.output, c.input, url)
			}
			if err == nil && c.err {
				t.Errorf("expected an error but got none")
			}
			if err != nil && !c.err {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestPrettifyGitHubIssueURL(t *testing.T) {
	var cases = []struct {
		input  string
		mode   PrettyMode
		output string
	}{
		{"https://github.com/myorg/repo1/issues/123", PrettyModeNum, "#123"},
		{"https://github.com/myorg/repo1/issues/123", PrettyModeRepo, "repo1#123"},
		{"https://github.com/myorg/repo1/issues/123", PrettyModeOrgRepo, "myorg/repo1#123"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			url := PrettifyGitHubIssueURL(c.input, c.mode)
			if url != c.output {
				t.Errorf("expected <%s> for input <%s>, got <%s>", c.output, c.input, url)
			}
		})
	}
}
