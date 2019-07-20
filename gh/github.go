package gh

import (
	"context"
	"github.com/google/go-github/v27/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"strings"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewGitHubClient(accessToken string) Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	return Client{
		client: client,
		ctx:    ctx,
	}
}

func (c Client) EnableBranchProtection(owner string, repo string, branch string) error {
	request := &github.ProtectionRequest{
		RequiredStatusChecks: nil,
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
			DismissalRestrictionsRequest: &github.DismissalRestrictionsRequest{
				Users: &[]string{},
				Teams: &[]string{},
			},
			DismissStaleReviews:     false,
			RequireCodeOwnerReviews: false,
		},
		EnforceAdmins: true,
		Restrictions:  nil,
	}

	_, _, err := c.client.Repositories.UpdateBranchProtection(c.ctx, owner, repo, branch, request)

	return err
}

func (c Client) DisableBranchProtection(owner string, repo string, branch string) error {
	response, err := c.client.Repositories.RemoveBranchProtection(c.ctx, owner, repo, branch)

	if response.StatusCode == 404 {
		return nil
	}

	return err
}

func SplitRepoString(repository string) (string, string, error) {
	tokens := strings.Split(repository, "/")

	if len(tokens) != 2 {
		return "", "", errors.Errorf("Unable to split repository string into owner and repo name: %s", repository)
	}

	return tokens[0], tokens[1], nil
}
