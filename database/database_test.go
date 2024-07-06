package database_test

import (
	"context"
	"testing"

	"identity/config"
	"identity/database/postgres"
)

func TestNewPostgres(t *testing.T) {
	dbConn := postgres.New(
		context.Background(),
		&config.DatabaseConfig{
			Type:    "postgres",
			Uri:     "",
			Timeout: 30,
		},
	)
	if dbConn == nil {
		t.Error("Failed to create postgres connection")
	}
}
