package entities

import "time"

type Commit struct {
	ID         uint
	CommitHash string
	Message    string
	Date       time.Time
	Author     Author
}
