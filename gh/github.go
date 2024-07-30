package gh

import (
	"context"
	"fmt"
	"github.com/google/go-github/v63/github"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"slices"
	"strings"
)

var validRulesetEnforcements = []string{"active", "disabled"}

type Ruleset struct {
	ID          int64
	Name        string
	Enforcement string
	Owner       string
	Repo        string
}

type Client struct {
	client *github.Client
	ctx    context.Context
}

func (gh *Client) updateRulesetEnforcementByName(owner string, repo string, rulesetName string, enforcement string) (*Ruleset, error) {
	if !slices.Contains(validRulesetEnforcements, enforcement) {
		return nil, fmt.Errorf("invalid ruleset enforement value: %s (valid: %s)", enforcement, strings.Join(validRulesetEnforcements, ", "))
	}

	rulesets, _, err := gh.client.Repositories.GetAllRulesets(gh.ctx, owner, repo, false)

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve rulesets for %s/%s: %w", owner, repo, err)
	}

	ruleset, found := lo.Find(rulesets, func(item *github.Ruleset) bool {
		return item.Name == rulesetName
	})
	if !found {
		return nil, fmt.Errorf("couldn't find ruleset with name %q in rulesets: %s", rulesetName,
			strings.Join(lo.Map(rulesets, func(item *github.Ruleset, index int) string {
				return item.Name
			}), ", "))
	}

	rs := &Ruleset{
		ID:          ruleset.GetID(),
		Name:        ruleset.Name,
		Enforcement: ruleset.Enforcement,
		Owner:       owner,
		Repo:        repo,
	}

	if ruleset.Enforcement == enforcement {
		// Nothing to do
		return rs, nil
	}

	ruleset.Enforcement = enforcement
	rs.Enforcement = enforcement

	if _, _, err := gh.client.Repositories.UpdateRuleset(gh.ctx, owner, repo, ruleset.GetID(), ruleset); err != nil {
		return nil, fmt.Errorf("couldn't update ruleset %q in repo %s/%s: %w", ruleset.Name, owner, repo, err)
	}

	return rs, nil
}

func (gh *Client) EnableRulesetByName(owner string, repo string, rulesetName string) (*Ruleset, error) {
	return gh.updateRulesetEnforcementByName(owner, repo, rulesetName, "active")
}

func (gh *Client) DisableRulesetByName(owner string, repo string, rulesetName string) (*Ruleset, error) {
	return gh.updateRulesetEnforcementByName(owner, repo, rulesetName, "disabled")
}

func NewGitHubClient(accessToken string) *Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func SplitRepoString(repository string) (string, string, error) {
	tokens := strings.Split(repository, "/")

	if len(tokens) != 2 {
		return "", "", errors.Errorf("Unable to split repository string into owner and repo name: %s", repository)
	}

	return tokens[0], strings.TrimSuffix(tokens[1], ".git"), nil
}
