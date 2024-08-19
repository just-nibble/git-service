package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.github.com"

// GitHubClient is a simple client for interacting with GitHub's API
type GitHubClient struct {
	HTTPClient *http.Client
}

// NewGitHubClient creates a new instance of GitHubClient with a timeout
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Repository represents the JSON structure of a GitHub repository
type Repository struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"html_url"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	ForksCount      int `json:"forks_count"`
	StarsCount      int `json:"stargazers_count"`
	OpenIssuesCount int `json:"open_issues_count"`
	WatchersCount   int `json:"watchers_count"`
}

// Commit represents the JSON structure of a GitHub commit
type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		URL string `json:"url"`
	} `json:"commit"`
}

// GetRepository fetches details of a GitHub repository by its owner and name
func (c *GitHubClient) GetRepository(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", baseURL, owner, repo)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repository: received status code %d", resp.StatusCode)
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode repository response: %v", err)
	}

	return &repository, nil
}

// GetCommits fetches the commits for a GitHub repository by its owner, name, and since date
// Handles pagination by following the "Link" header
func (c *GitHubClient) GetCommits(owner, repo string, since time.Time) ([]Commit, error) {
	var allCommits []Commit
	url := fmt.Sprintf("%s/repos/%s/%s/commits?since=%s", baseURL, owner, repo, since.Format(time.RFC3339))

	if url != "" {
		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commits: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch commits: received status code %d", resp.StatusCode)
		}

		var commits []Commit
		if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
			return nil, fmt.Errorf("failed to decode commits response: %v", err)
		}

		allCommits = append(allCommits, commits...)
	}

	return allCommits, nil
}

// extractNextLink parses the Link header to find the "next" URL
func extractNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) >= 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
			return strings.TrimSpace(parts[0][1 : len(parts[0])-1])
		}
	}

	return ""
}
