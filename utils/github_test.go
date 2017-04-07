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
}
