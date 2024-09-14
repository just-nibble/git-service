package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/pkg/api"
	"github.com/just-nibble/git-service/pkg/log"
)

type GitHubClient struct {
	baseURL       string
	token         string
	fetchInterval time.Duration
	client        *api.RestClient
	rateLimit     RateLimit
	log           log.Log
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int64
}

type GitHubCommitResponse struct {
	SHA     string `json:"sha"`
	NodeID  string `json:"node_id"`
	Commit  Commit `json:"commit"`
	Author  Author `json:"author"`
	HtmlURL string `json:"html_url"`
}

type Commit struct {
	Author  Author `json:"author"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

type Author struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

type GitHubMetaResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HtmlURL     string `json:"html_url"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Owner       struct {
		URL     string `json:"url"`
		HtmlURL string `json:"html_url"`
	} `json:"owner"`
	StargazersCount int    `json:"stargazers_count"`
	WatchersCount   int    `json:"watchers_count"`
	Language        string `json:"language"`
	ForksCount      int    `json:"forks_count"`
	OpenIssues      int    `json:"open_issues"`
}

type GithubPaging struct {
	Limit     int    `json:"limit,omitempty"`
	Page      int    `json:"page,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type GithubPagingInfo struct {
	TotalCount  int64 `json:"totalCount"`
	Page        int   `json:"page"`
	HasNextPage bool  `json:"hasNextPage"`
	Count       int   `json:"count"`
}

// NewGitHubClient creates a new instance of GitHubClient.
func NewGitHubClient(baseURL, token string, fetchInterval time.Duration) GitClient {
	client := api.NewRestClient()

	return &GitHubClient{
		baseURL:       baseURL,
		token:         token,
		fetchInterval: fetchInterval,
		client:        client,
	}
}

// FetchRepoMetadata fetches repository metadata from GitHub.
func (g *GitHubClient) FetchRepoMetadata(ctx context.Context, repositoryName string) (*domain.RepositoryMeta, error) {
	endpoint := fmt.Sprintf("https://%s/repos/%s", g.baseURL, repositoryName)

	resp, err := g.client.Get(endpoint, nil, g.getHeaders())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository metadata: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	var gitHubRepo GitHubMetaResponse
	if err := json.Unmarshal([]byte(resp.Body), &gitHubRepo); err != nil {
		g.log.Error.Println("failed to unmarshal repository metadata response")
		return nil, errors.New("failed to parse repository metadata response")
	}

	return &domain.RepositoryMeta{
		Name:            gitHubRepo.FullName,
		Description:     gitHubRepo.Description,
		URL:             gitHubRepo.URL,
		Language:        gitHubRepo.Language,
		ForksCount:      gitHubRepo.ForksCount,
		StarsCount:      gitHubRepo.StargazersCount,
		OpenIssuesCount: gitHubRepo.OpenIssues,
		WatchersCount:   gitHubRepo.WatchersCount,
	}, nil
}

// FetchCommits fetches a list of commits from GitHub.
func (g *GitHubClient) FetchCommits(ctx context.Context, repo domain.RepositoryMeta, since, until time.Time, lastFetchedCommit string, page, perPage int) ([]domain.Commit, bool, error) {
	endpoint, err := g.buildCommitEndpoint(repo.Name, since, until, lastFetchedCommit, page, perPage)
	if err != nil {
		return nil, false, fmt.Errorf("failed to build commit endpoint: %w", err)
	}

	resp, err := g.client.Get(endpoint, nil, g.getHeaders())
	if err != nil {
		g.log.Error.Println("failed to fetch commits")
		return nil, false, fmt.Errorf("failed to fetch commits: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, false, fmt.Errorf("rate limit exceeded")
	}

	g.updateRateLimit(resp)

	if g.rateLimit.Remaining == 0 {
		waitTime := time.Until(time.Unix(g.rateLimit.Reset, 0))
		g.log.Info.Printf("Rate limit exceeded. Waiting for %v until reset...", waitTime)
		time.Sleep(waitTime)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	var commitRes []GitHubCommitResponse
	if err := json.Unmarshal([]byte(resp.Body), &commitRes); err != nil {
		g.log.Error.Println("failed to unmarshal commits response")
		return nil, false, errors.New("failed to parse commits response")
	}

	commits := g.parseCommits(commitRes, repo.Name)

	// Determine if there are more pages to fetch
	morePages := g.hasNextPage(resp.Headers["Link"])

	return commits, morePages, nil
}

func (g *GitHubClient) buildCommitEndpoint(repoName string, since, until time.Time, lastFetchedCommit string, page, perPage int) (string, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s/repos/%s/commits", g.baseURL, repoName))
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	if lastFetchedCommit != "" {
		q.Set("sha", lastFetchedCommit)
	} else {
		q.Set("since", since.Format(time.RFC3339))
		q.Set("until", until.Format(time.RFC3339))
	}
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("page", strconv.Itoa(page))

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (g *GitHubClient) parseCommits(commitRes []GitHubCommitResponse, repoName string) []domain.Commit {
	commits := make([]domain.Commit, len(commitRes))
	for i, cr := range commitRes {
		commits[i] = domain.Commit{
			Hash:    cr.SHA,
			Message: cr.Commit.Message,
			Author: domain.Author{
				Name:  cr.Commit.Author.Name,
				Email: cr.Commit.Author.Email,
			},
			Date: cr.Commit.Author.Date,
		}
	}
	return commits
}

// hasNextPage checks if there is a 'next' link in the Link header.
func (g *GitHubClient) hasNextPage(linkHeader []string) bool {
	if len(linkHeader) == 0 {
		return false
	}
	links := g.parseLinkHeader(linkHeader[0])
	_, hasNext := links["next"]
	return hasNext
}

// parseLinkHeader parses the Link header into a map of rel to URLs.
func (g *GitHubClient) parseLinkHeader(header string) map[string]string {
	links := make(map[string]string)
	parts := strings.Split(header, ",")
	for _, part := range parts {
		sections := strings.Split(part, ";")
		if len(sections) < 2 {
			continue
		}
		url := strings.Trim(sections[0], " <>")
		rel := strings.Trim(sections[1], ` rel="`)
		links[rel] = url
	}
	return links
}

// updateRateLimit updates the rate limit values from the response headers.
func (g *GitHubClient) updateRateLimit(resp *api.HTTPResponse) {
	g.rateLimit.Limit = parseHeaderInt(resp.Headers, "X-Ratelimit-Limit")
	g.rateLimit.Remaining = parseHeaderInt(resp.Headers, "X-Ratelimit-Remaining")
	g.rateLimit.Reset = parseHeaderInt64(resp.Headers, "X-Ratelimit-Reset")
}

func (g *GitHubClient) getHeaders() map[string]string {
	if len(g.token) == 0 {
		return map[string]string{}
	}
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", g.token),
	}
}

func parseHeaderInt(headers map[string][]string, key string) int {
	if vals, ok := headers[key]; ok && len(vals) > 0 {
		val, err := strconv.Atoi(vals[0])
		if err == nil {
			return val
		}
	}
	return 0
}

func parseHeaderInt64(headers map[string][]string, key string) int64 {
	if vals, ok := headers[key]; ok && len(vals) > 0 {
		val, err := strconv.ParseInt(vals[0], 10, 64)
		if err == nil {
			return val
		}
	}
	return 0
}
