package dtos

import "time"

type RepositoryInput struct {
	Name string `json:"name"`
}

// Repository represents the JSON structure of a GitHub repository
type RepositoryMeta struct {
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
