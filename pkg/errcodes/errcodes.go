package errcodes

import "errors"

var (
	ErrNoRecordFound            = errors.New("no record found")
	ErrInvalidInput             = errors.New("invalid input")
	ErrInvalidRepositoryId      = errors.New("invalid repository ID")
	ErrResolvingRepositoryName  = errors.New("no repository meta data was found with specified name")
	ErrDefaultRepoAlreadySeeded = errors.New("default repo already seeded")
	ErrRepoAlreadyAdded         = errors.New("repository is already added")

	ErrRepoMetaDataNotFetched = errors.New("repository metadata not fetched, ensure repository is valid and public")
	ErrInvalidRepositoryName  = errors.New("invalid repository name, eg format is {owner/repositoryName}")

	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrContextCancelled  = errors.New("context cancelled")
)
