package dtos

// Commit represents the JSON structure of a GitHub commit
type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  Author `json:"author"`
		URL     string `json:"url"`
	} `json:"commit"`
}

type MultiCommitsResponse struct {
	Commits  []Commit   `json:"commits"`
	PageInfo PagingInfo `json:"page_info"`
}
