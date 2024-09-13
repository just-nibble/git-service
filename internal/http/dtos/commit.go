package dtos

import "time"

// Commit represents the JSON structure of a GitHub commit
type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  Author `json:"author"`
		URL     string `json:"url"`
	} `json:"commit"`
}

type CommitReponse struct {
	ID       uint      `json:"id"`
	Hash     string    `json:"hash"`
	Message  string    `json:"message"`
	Date     time.Time `json:"date"`
	Author   Author    `json:"author"`
	AuthorID uint      `json:"author_id"`
	RepoID   uint      `json:"repo_id"`
}

type MultiCommitsResponse struct {
	Commits  []CommitReponse `json:"commits"`
	PageInfo PagingInfo      `json:"page_info"`
}
