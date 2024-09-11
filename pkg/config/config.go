package config

import (
	"time"

	"github.com/just-nibble/git-service/internal/adapters/validators"
)

type Config struct {
	DefaultRepository validators.Repo `validate:"required"`
	DefaultStartDate  time.Time
	MonitorInterval   int
}
