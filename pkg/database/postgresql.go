package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/just-nibble/git-service/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresDatabase struct {
	Dsn             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	db              *gorm.DB
}

func NewPostgresDatabase(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) *PostgresDatabase {
	return &PostgresDatabase{
		Dsn:             dsn,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}
}

// ConnectDb establishes the Postgres database connection.
func (p *PostgresDatabase) ConnectDB(ctx context.Context) error {
	db, err := gorm.Open(postgres.Open(p.Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(p.MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(p.ConnMaxLifetime)

	p.db = db

	log.Println("Postgres database connected successfully")
	return nil
}

// GetDB returns the underlying *gorm.DB instance.
func (p *PostgresDatabase) GetDB() *gorm.DB {
	return p.db
}

// Migrate handles schema migrations for Postgres.
func (p *PostgresDatabase) Migrate(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}

	// Assuming models like User, Product, etc.
	if err := p.db.AutoMigrate(&repository.Author{}, &repository.Repository{}, &repository.Commit{}); err != nil {
		return fmt.Errorf("failed to migrate postgres: %w", err)
	}

	log.Println("Postgres migrations applied successfully")
	return nil
}

// PingDb checks if the Postgres database connection is alive.
func (p *PostgresDatabase) PingDb(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}

	log.Println("Postgres database connection is alive")
	return nil
}

// CloseDb closes the Postgres database connection.
func (p *PostgresDatabase) CloseDb(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close postgres database: %w", err)
	}

	log.Println("Postgres database closed successfully")
	return nil
}
