package storage

import (
	"fmt"
	"log"
	"os"

	"github.com/just-nibble/git-service/internal/core/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// Set up PostgreSQL connection details
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Automatically migrate the schema
	if err := db.AutoMigrate(&entities.Repository{}, &entities.Commit{}, &entities.Author{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
