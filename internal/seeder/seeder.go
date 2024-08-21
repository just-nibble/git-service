package seeder

import (
	"log"

	"github.com/just-nibble/git-service/internal/data"
	"github.com/just-nibble/git-service/internal/service"
	"gorm.io/gorm"
)

// SeedDatabase seeds the database with the Chromium repository, along with its commits and authors, if the database is empty
func SeedDatabase(db *gorm.DB, Indexer *service.Indexer) error {
	// Check if the repository table is empty
	var count int64
	if err := db.Model(&data.Repository{}).Count(&count).Error; err != nil {
		return err
	}

	// If the repository table is empty, seed with the Chromium repository
	if count == 0 {
		log.Println("Seeding database with Chromium repository...")

		// Fetch repository details from GitHub
		repo, err := Indexer.GetRepository("chromium", "chromium")
		if err != nil {
			log.Println(err)
			return err
		}

		// Save repository in the database
		chromiumRepo := &data.Repository{
			OwnerName:       repo.Owner.Login,
			Name:            repo.Name,
			URL:             repo.URL,
			ForksCount:      repo.ForksCount,
			StarsCount:      repo.StarsCount,
			OpenIssuesCount: repo.OpenIssuesCount,
			WatchersCount:   repo.WatchersCount,
		}

		if err := db.Create(&chromiumRepo).Error; err != nil {
			return err
		}

		// Start background indexing of commits
		go func() {
			if err := Indexer.IndexCommits(chromiumRepo); err != nil {
				log.Println(err)
			}
		}()

		log.Println("Database seeding completed with Chromium repository, commits, and authors.")
	}

	return nil
}
