package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.github.com"

// GitHubClient is a simple client for interacting with GitHub's API
type GitHubClient struct {
	HTTPClient    *http.Client
	Authorization string
}

// NewGitHubClient creates a new instance of GitHubClient with a timeout
func NewGitHubClient(token string) *GitHubClient {
	client := &GitHubClient{
		HTTPClient:    http.DefaultClient,
		Authorization: token,
	}
	return client
}

// Repository represents the JSON structure of a GitHub repository
type Repository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"html_url"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	ForksCount      int       `json:"forks_count"`
	StarsCount      int       `json:"stargazers_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	WatchersCount   int       `json:"watchers_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if c.Authorization != "" {
			req.Header.Add("Authorization", "token "+c.Authorization)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repository: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var repository Repository
			if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
				return nil, fmt.Errorf("failed to decode repository response: %w", err)
			}
			return &repository, nil
		}

		if resp.StatusCode == http.StatusForbidden {
			// Check if rate limiting is the cause
			if resetTime := resp.Header.Get("X-RateLimit-Reset"); resetTime != "" {
				resetTimestamp, err := strconv.ParseInt(resetTime, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse X-RateLimit-Reset header: %w", err)
				}

				sleepDuration := time.Until(time.Unix(resetTimestamp, 0))
				log.Printf("Rate limit exceeded. Sleeping for %v\n", sleepDuration)
				time.Sleep(sleepDuration)
				continue
			}
			return nil, fmt.Errorf("rate limit exceeded or forbidden: received status code %d", resp.StatusCode)
		}

		return nil, fmt.Errorf("failed to fetch repository: received status code %d", resp.StatusCode)
	}
}

// GetCommits fetches the commits for a GitHub repository by its owner, name, and since date
func (c *GitHubClient) GetCommits(owner, repo string, since time.Time, page int, perPage int) ([]Commit, error) {
	url := fmt.Sprintf(
		"%s/repos/%s/%s/commits?since=%s&page=%d&per_page=%d",
		baseURL, owner, repo, since.Format(time.RFC3339),
		page, perPage,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Authorization != "" {
		req.Header.Add("Authorization", "token "+c.Authorization)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return nil, errors.New("rate limited") // Return rate-limited status
		}
		return nil, fmt.Errorf("failed to fetch commits: received status code %d", resp.StatusCode)
	}

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, fmt.Errorf("failed to decode commits response: %w", err)
	}

	return commits, nil
}
