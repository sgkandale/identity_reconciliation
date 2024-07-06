package postgres

import (
	"context"
	"log"
	"strings"
	"time"

	"identity/config"
	"identity/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	*pgxpool.Pool
	timeout time.Duration
}

const (
	TableName_Contact = "Contact"
)

func New(ctx context.Context, cfg *config.DatabaseConfig) database.Database {
	log.Print("[INFO] creating postgres connection")
	if cfg == nil {
		log.Fatal("[ERROR] postgres config is nil")
	}
	if !strings.EqualFold(cfg.Type, "postgres") {
		log.Fatal("[ERROR] postgres config type is not 'postgres'")
	}

	timeout := time.Second * time.Duration(cfg.Timeout)
	connectCtx, cancelConnectCtx := context.WithTimeout(ctx, timeout)
	defer cancelConnectCtx()

	pgxCfg, err := pgxpool.ParseConfig(cfg.Uri)
	if err != nil {
		log.Fatal("[ERROR] postgres config parse error: ", err)
	}

	pgxPool, err := pgxpool.NewWithConfig(
		connectCtx,
		pgxCfg,
	)
	if err != nil {
		log.Fatal("[ERROR] postgres connect error: ", err)
	}

	pingCtx, cancelPingCtx := context.WithTimeout(ctx, timeout)
	defer cancelPingCtx()

	log.Print("[INFO] pinging postgres")
	err = pgxPool.Ping(pingCtx)
	if err != nil {
		log.Fatal("[ERROR] postgres ping error: ", err)
	}

	log.Print("[INFO] postgres connection created")

	return &Client{
		Pool:    pgxPool,
		timeout: timeout,
	}
}
