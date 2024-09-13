package dtos

import "time"

type Author struct {
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Date        time.Time `json:"date"`
	CommitCount int       `json:"commit_count"`
}
