package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/11SF/go-common/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type PostgresBunConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresConnectionWithBun(config PostgresBunConfig) *bun.DB {

	ctx := context.Background()
	logger.Info(ctx, "connecting to database", logger.LogAttrTag("postgres setup"))

	tlsSkipVerify := config.SSLMode == "disable"
	pgconn := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(fmt.Sprintf("%s:%s", config.Host, config.Port)),
		pgdriver.WithDatabase(config.DBName),
		pgdriver.WithUser(config.Username),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithInsecure(tlsSkipVerify),
	)

	sqldb := sql.OpenDB(pgconn)
	bunDB := bun.NewDB(sqldb, pgdialect.New())

	err := bunDB.Ping()
	if err != nil {
		slog.Error(err.Error())
		panic(0)
	}

	logger.Info(ctx, "Database connected", logger.LogAttrTag("postgres setup"))
	return bunDB
}
