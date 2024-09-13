package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/internal/http/dtos"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GormCommitStore is a GORM-based implementation of CommitStore
type GormCommitStore struct {
	db *gorm.DB
}

// NewGormCommitStore initializes a new GormCommitStore
func NewGormCommitStore(db *gorm.DB) CommitStore {
	return &GormCommitStore{db: db}
}

func (s *GormCommitStore) GetCommitByHash(ctx context.Context, hash string) (*domain.Commit, error) {
	if ctx.Err() == context.Canceled {
		return nil, errcodes.ErrContextCancelled
	}
	var commit Commit
	err := s.db.WithContext(ctx).Where("commit_hash = ?", hash).Find(&commit).Error

	if commit.ID == 0 {
		return nil, errcodes.ErrNoRecordFound
	}
	return commit.ToDomain(), err
}

// SaveCommit stores a repository commit into the database
func (s *GormCommitStore) SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error) {
	if ctx.Err() == context.Canceled {
		return nil, errcodes.ErrContextCancelled
	}

	author := Author{}

	s.db.WithContext(ctx).Where(&Author{
		Name:  commit.Author.Name,
		Email: commit.Author.Email,
	}).FirstOrCreate(&author)

	commit.AuthorID = author.ID

	dbCommit := ToGormCommit(&commit)

	tx := s.db.WithContext(ctx).Create(&dbCommit)

	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), `duplicate key value violates unique constraint`) {
			return nil, tx.Error
		}
		return nil, tx.Error
	}
	return dbCommit.ToDomain(), nil
}

// GetAllCommitsByRepositoryName fetches all stores commits by repository name
func (s *GormCommitStore) GetCommitsByRepository(ctx context.Context, repo domain.RepositoryMeta, query dtos.APIPagingDto) (*dtos.MultiCommitsResponse, error) {
	var dbCommits []Commit

	var count, queryCount int64

	queryInfo, offset := getPaginationInfo(query)

	db := s.db.WithContext(ctx).Model(&Commit{}).Where(&Commit{RepositoryID: repo.ID})

	db.Count(&count)

	db = db.Offset(offset).Limit(queryInfo.Limit).
		Order(fmt.Sprintf("commit.%s %s", queryInfo.Sort, queryInfo.Direction)).
		Find(&dbCommits)
	db.Count(&queryCount)

	if db.Error != nil {
		log.Info().Msgf("fetch commits error %v", db.Error.Error())

		return nil, db.Error
	}

	pagingInfo := getPagingInfo(queryInfo, int(count))
	pagingInfo.Count = len(dbCommits)

	return &dtos.MultiCommitsResponse{
		Commits:  commitResponse(dbCommits),
		PageInfo: pagingInfo,
	}, nil

}

func commitResponse(commits []Commit) []dtos.Commit {
	if len(commits) == 0 {
		return nil
	}

	commitsResponse := make([]dtos.Commit, 0, len(commits))

	for _, c := range commits {
		cr := dtos.Commit{
			SHA: c.CommitHash,
			Commit: struct {
				Message string      "json:\"message\""
				Author  dtos.Author "json:\"author\""
				URL     string      "json:\"url\""
			}{
				Message: c.Message,
				Author: dtos.Author{
					Name:  c.Author.Name,
					Email: c.Author.Email,
					Date:  c.Date,
				},
			},
		}

		commitsResponse = append(commitsResponse, cr)
	}

	return commitsResponse
}
