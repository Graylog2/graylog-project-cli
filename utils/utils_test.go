package utils_test

import (
	"github.com/Graylog2/graylog-project-cli/utils"
	"testing"
)

const sshRepo = "git@github.com:Graylog2/graylog2-server.git"
const httpsRepo = "https://github.com/Graylog2/graylog2-server.git"

var repos = [...]string {
	"git@github.com:Graylog2/graylog2-server.git",
	"https://github.com/Graylog2/graylog2-server.git",
	"user@example.org:External/Graylog2/graylog2-server.git",
	"https://example.org/External/Graylog2/graylog2-server.git",
	"file:///home/user/Graylog2/graylog2-server",
	"git@github.com:Graylog2/graylog2-server/",
	"user@example.org:graylog2-server",
	"user@example.org:graylog2-server/",
	"../../graylog2-server",
}

func TestNameFromRepository(t *testing.T) {
	expected := "graylog2-server"

	sshRepoName := utils.NameFromRepository(sshRepo)
	if sshRepoName != expected {
		t.Errorf("Repository %s does not resolve to name %s - result: `%s`", sshRepo, expected, sshRepoName)
	}

	httpsRepoName := utils.NameFromRepository(httpsRepo)
	if httpsRepoName != expected {
		t.Errorf("Repository %s does not resolve to name %s - result: `%s`", httpsRepo, expected, httpsRepoName)
	}


	for _, repo := range repos {
		repoName := utils.NameFromRepository(repo)
		if repoName != expected {
			t.Errorf("Repository %s does not resolve to name %s - result: `%s`", repo, expected, repoName)
		}
	}
}

func TestConvertGithubGitToHTTPS(t *testing.T) {
	toHTTPS := utils.ConvertGithubGitToHTTPS(sshRepo)
	if toHTTPS != httpsRepo {
		t.Errorf("Repository %s was not converted to %s - result: %s", sshRepo, httpsRepo, toHTTPS)
	}

	toHTTPS = utils.ConvertGithubGitToHTTPS(httpsRepo)
	if toHTTPS != httpsRepo {
		t.Errorf("Repository %s was converted to %s - that should not happen", httpsRepo, toHTTPS)
	}
}
