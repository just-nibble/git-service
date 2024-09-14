package errcodes

import "errors"

var (
	// General Errors
	ErrNoRecordFound    = errors.New("no record found")
	ErrContextCancelled = errors.New("operation cancelled by context")

	// Repository Errors
	ErrRepoAlreadyAdded      = errors.New("repository has already been added")
	ErrInvalidRepositoryName = errors.New("invalid repository name, expected format: {owner/repositoryName}")
)
