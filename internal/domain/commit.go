package domain

import "time"

type Commit struct {
	ID       uint
	Hash     string
	Message  string
	Date     time.Time
	Author   Author
	AuthorID uint
	RepoID   uint
}
