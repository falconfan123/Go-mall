package db

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/services/users/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "github.com/lib/pq"
)

func NewPostgres(postgresConf config.PostgresConfig) sqlx.SqlConn {
	// 使用 postgres 驱动连接 PostgreSQL
	conn := sqlx.NewSqlConn("postgres", postgresConf.DataSource)
	db, err := conn.RawDB()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(postgresConf.Conntimeout))
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {

		panic(err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	return conn

}
