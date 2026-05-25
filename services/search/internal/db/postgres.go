package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/falconfan123/Go-mall/services/search/internal/config"

	_ "github.com/lib/pq"
)

func NewPostgres(postgresConf config.PostgresConfig) *sql.DB {
	db, err := sql.Open("postgres", postgresConf.DataSource)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(postgresConf.Conntimeout))
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	return db
}
