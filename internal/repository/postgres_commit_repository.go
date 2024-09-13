package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/just-nibble/git-service/internal/domain"
	"github.com/just-nibble/git-service/pkg/errcodes"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GormCommitRepository is a GORM-based implementation of CommitRepository
type GormCommitRepository struct {
	db *gorm.DB
}

// NewGormCommitRepository initializes a new GormCommitRepository
func NewGormCommitRepository(db *gorm.DB) CommitRepository {
	return &GormCommitRepository{db: db}
}

func (s *GormCommitRepository) GetCommitByHash(ctx context.Context, hash string) (*domain.Commit, error) {
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
func (s *GormCommitRepository) SaveCommit(ctx context.Context, commit domain.Commit) (*domain.Commit, error) {
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
func (s *GormCommitRepository) GetCommitsByRepository(ctx context.Context, repo domain.RepositoryMeta, query domain.APIPaging) ([]domain.Commit, error) {
	var dbCommits []Commit

	var count, queryCount int64

	queryInfo, offset := getPaginationInfo(query)

	db := s.db.WithContext(ctx).Model(&Commit{}).Where(&Commit{RepositoryID: repo.ID})

	db.Count(&count)

	db = db.Offset(offset).Limit(queryInfo.Limit).
		Order(fmt.Sprintf("commit.%s %s", queryInfo.Sort, queryInfo.Direction)).
		Preload("Author").Find(&dbCommits)
	db.Count(&queryCount)

	if db.Error != nil {
		log.Info().Msgf("fetch commits error %v", db.Error.Error())

		return nil, db.Error
	}

	pagingInfo := getPagingInfo(queryInfo, int(count))
	pagingInfo.Count = len(dbCommits)

	var commits []domain.Commit

	for _, commit := range dbCommits {
		v := domain.Commit{
			ID:      commit.ID,
			Hash:    commit.CommitHash,
			Message: commit.Message,
			Author: domain.Author{
				ID:          commit.AuthorID,
				Name:        commit.Author.Name,
				Email:       commit.Author.Email,
				CommitCount: commit.Author.CommitCount,
			},
			RepoID: commit.RepositoryID,
		}
		commits = append(commits, v)
	}

	return commits, nil

}
