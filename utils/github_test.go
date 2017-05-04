package utils

import "testing"

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

	// Missing .git suffix
	_, err = ParseGitHubURL("https://github.com/Graylog2/graylog2-server")
	if err == nil {
		t.Error("expected URL without .git suffix to fail")
	}

	// Unknown URL format
	_, err = ParseGitHubURL("https://example.com/Graylog2/graylog2-server")
	if err == nil {
		t.Error("expected unknown URL to fail")
	}

	url, _ = ParseGitHubURL("github://Graylog2/graylog2-server.git")
	if url.IsHTTPS() || url.IsHTTPS() {
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
