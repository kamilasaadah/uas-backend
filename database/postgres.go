package database

import (
	"context"
	"log"

	"uas-backend/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PG *pgxpool.Pool

func ConnectPostgres() {
	dsn := config.DBDsn()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect PostgreSQL: %v", err)
	}

	PG = pool
	log.Println("✅ Connected to PostgreSQL")
}
