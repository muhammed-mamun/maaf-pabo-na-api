package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

type Client struct {
	client *github.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	client := github.NewClient(nil)

	return &Client{client: client}, nil
}

func (c *Client) GetUser(ctx context.Context, username string) (*github.User, error) {
	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (c *Client) GetRepositories(ctx context.Context, username string) ([]*github.Repository, error) {
	// Fixing the typo here, change "serchOpts" to "searchOpts"
	searchOpts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Search.Repositories(ctx, fmt.Sprintf("user:%s", username), searchOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to get repositories: %w", err)
		}
		allRepos = append(allRepos, repos.Repositories...)

		// Check if there are more pages of repositories to fetch
		if resp.NextPage == 0 {
			break
		}
		searchOpts.Page = resp.NextPage
	}
	return allRepos, nil
}

func ExtractRepoNames(repos []*github.Repository) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.GetName()
	}
	return names
}
