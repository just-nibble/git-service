package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
func (c *GitHubClient) GetCommits(owner, repo string, since time.Time, page int, perPage int) ([]Commit, bool, error) {
	url := fmt.Sprintf(
		"%s/repos/%s/%s/commits?since=%s&page=%d&per_page=%d",
		baseURL, owner, repo, since.Format(time.RFC3339),
		page, perPage,
	)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch commits: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate-limiting scenario
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, true, nil // Return rate-limited status
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("failed to fetch commits: received status code %d", resp.StatusCode)
	}

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, false, fmt.Errorf("failed to decode commits response: %v", err)
	}

	return commits, false, nil
}

// fetchPage fetches a single page of commits from GitHub and returns whether the request was rate-limited
func (c *GitHubClient) FetchPage(owner, repo string, page, perPage int) ([]Commit, bool, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/repos/%s/%s/commits?page=%d&per_page=%d", baseURL, owner, repo, page, perPage)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if len(token) != 0 && token != "Not a real token" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate-limiting scenario
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, true, nil // Return rate-limited status
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, false, fmt.Errorf("failed to decode commits response: %v", err)
	}

	return commits, false, nil
}
