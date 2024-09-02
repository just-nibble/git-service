package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// MockTransport is a mock implementation of http.RoundTripper for testing purposes
type MockTransport struct {
	RoundTripper func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripper(req)
}

func TestGetRepositorySuccess(t *testing.T) {
	// Mock HTTP client with a successful response
	mockTransport := &MockTransport{
		RoundTripper: func(req *http.Request) (*http.Response, error) {
			expectedURL := fmt.Sprintf("%s/repos/%s/%s", baseURL, "octocat", "hello-world")
			if req.URL.String() != expectedURL {
				t.Errorf("Unexpected request URL: %s", req.URL.String())
				return nil, fmt.Errorf("unexpected request")
			}

			// Simulate a successful response with repository details
			responseBody, _ := json.Marshal(Repository{
				ID:   1,
				Name: "hello-world",
				URL:  "https://github.com/octocat/hello-world",
				Owner: struct {
					Login string `json:"login"`
				}{Login: "octocat"},
				ForksCount:      42,
				StarsCount:      100,
				OpenIssuesCount: 5,
				WatchersCount:   200,
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	// Create a GitHubClient with the mock client
	client := &GitHubClient{
		HTTPClient: &http.Client{
			Transport: mockTransport,
		},
	}

	// Call GetRepository
	repo, err := client.GetRepository("octocat", "hello-world")

	// Check for errors
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify returned repository details
	if repo.Name != "hello-world" {
		t.Errorf("Expected repository name 'hello-world', got %s", repo.Name)
	}
	if repo.Owner.Login != "octocat" {
		t.Errorf("Expected owner 'octocat', got %s", repo.Owner.Login)
	}
}

func TestGetCommitsSuccess(t *testing.T) {
	// Mock HTTP client with a successful response
	page := 1
	perPage := 100

	mockTransport := &MockTransport{
		RoundTripper: func(req *http.Request) (*http.Response, error) {
			expectedSince := "2012-03-06T23:06:50Z" // RFC3339 format
			expectedURL := fmt.Sprintf("%s/repos/%s/%s/commits?since=%s&page=%d&per_page=%d", baseURL, "octocat", "hello-world", expectedSince, page, perPage)

			if req.URL.String() != expectedURL {
				t.Logf("Expected URL: %s", expectedURL)
				t.Errorf("Unexpected request URL: %s", req.URL.String())
				return nil, fmt.Errorf("unexpected request")
			}

			// Simulate a successful response with a few commits
			responseBody := []byte(`[
				{
					"sha": "commit-sha-1",
					"commit": {
						"message": "First commit",
						"author": {
							"name": "John Doe",
							"email": "john.doe@example.com",
							"date": "2024-08-19T10:00:00Z"
						},
						"url": "https://github.com/octocat/hello-world/commit/commit-sha-1"
					}
				},
				{
					"sha": "commit-sha-2",
					"commit": {
						"message": "Second commit",
						"author": {
							"name": "Jane Doe",
							"email": "jane.doe@example.com",
							"date": "2024-08-18T10:00:00Z"
						},
						"url": "https://github.com/octocat/hello-world/commit/commit-sha-2"
					}
				}
			]`)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	// Create a GitHubClient with the mock client
	client := &GitHubClient{
		HTTPClient: &http.Client{
			Transport: mockTransport,
		},
	}

	// Parse the 'since' date
	since, err := time.Parse(time.RFC3339, "2012-03-06T23:06:50Z")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Call GetCommits
	commits, _, err := client.GetCommits("octocat", "hello-world", since, page, perPage)

	// Check for errors
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify returned commits
	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	} else {
		// Check individual commits
		if commits[0].Commit.Message != "First commit" {
			t.Errorf("Expected first commit message 'First commit', got %s", commits[0].Commit.Message)
		}
		if commits[1].Commit.Message != "Second commit" {
			t.Errorf("Expected second commit message 'Second commit', got %s", commits[1].Commit.Message)
		}
	}
}
