package database

import (
	"context"
)

type Database interface {
	ConnectDB(ctx context.Context) error
	Migrate(ctx context.Context) error
	PingDb(ctx context.Context) error
	CloseDb(ctx context.Context) error
	GetDB() interface{}
}
