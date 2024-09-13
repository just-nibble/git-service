package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/just-nibble/git-service/pkg/env"
	"github.com/just-nibble/git-service/pkg/log"
)

type Config struct {
	DefaultRepository     string
	DefaultStartDate      time.Time
	DefaultEndDate        time.Time
	MonitorInterval       time.Duration
	DBHost                string `validate:"required"`
	DBUser                string `validate:"required"`
	DBPassword            string `validate:"required"`
	DBName                string `validate:"required"`
	DBPort                uint   `validate:"required"`
	SSLMode               string `validate:"required"`
	DSN                   string
	GitClientToken        string
	GitClientBaseURL      string
	GitCommitFetchPerPage int
	ServerAddress         string
	ServerPort            string
}

func LoadConfig(log log.Log) (*Config, error) {
	var err error

	interval := env.Getenv("MONITOR_INTERVAL", "1h")
	if interval == "" {
		interval = "1h"
	}

	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		log.Error.Printf("Invalid MONITOR_INTERVAL :[%s] env format: %s", interval, err.Error())
		return nil, err
	}

	var sDate time.Time
	var eDate time.Time

	startDate := os.Getenv("DEFAULT_START_DATE")
	if startDate == "" {
		sDate = time.Now().AddDate(0, -10, 0)
	} else {
		sDate, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			log.Error.Printf("Invalid DEFAULT_START_DATE [%s] env format: %s", startDate, err.Error())
			return nil, err
		}
	}

	perPage := env.Getenv("GIT_COMMIT_FETCH_PER_PAGE", "100")
	commitPerPage, err := strconv.Atoi(perPage)
	if err != nil {
		commitPerPage = 100
		log.Error.Printf("Invalid GIT_COMMIT_FETCH_PER_PAGE [%s] env format passed, setting to 100: %s", perPage, err.Error())
	}

	endDate := os.Getenv("DEFAULT_END_DATE")
	if endDate == "" {
		eDate = time.Now()
	} else {
		eDate, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			log.Error.Printf("Invalid DEFAULT_END_DATE [%s] env format: %s", endDate, err.Error())
			return nil, err
		}
	}

	dBPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Error.Printf("Invalid DB_PORT [%d] env format: %s", dBPort, err.Error())
		return nil, err
	}

	configVar := Config{
		GitClientToken:        os.Getenv("GITHUB_TOKEN"),
		DBHost:                os.Getenv("DB_HOST"),
		DBUser:                os.Getenv("DB_USER"),
		DBPassword:            os.Getenv("DB_PASSWORD"),
		DBName:                os.Getenv("DB_NAME"),
		DBPort:                uint(dBPort),
		SSLMode:               env.Getenv("DB_SSL_MODE", "disable"),
		MonitorInterval:       intervalDuration,
		DefaultStartDate:      sDate,
		DefaultEndDate:        eDate,
		GitCommitFetchPerPage: commitPerPage,
		GitClientBaseURL:      os.Getenv("GIT_API_BASE_URL"),
		ServerAddress:         env.Getenv("SERVER_ADDRESS", "localhost"),
		ServerPort:            env.Getenv("SERVER_PORT", "8080"),
		DefaultRepository:     env.Getenv("DEFAULT_REPOSITORY", "chromium/chromium"),
	}
	configVar.DSN = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		configVar.DBHost, configVar.DBUser, configVar.DBPassword, configVar.DBName, configVar.DBPort, configVar.SSLMode,
	)

	validate := validator.New()
	err = validate.Struct(configVar)
	if err != nil {
		log.Error.Printf("env validation error: %s", err.Error())
		return nil, err
	}

	return &configVar, nil
}
