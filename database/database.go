package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/11SF/go-common/logger"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"gorm.io/gorm"
)

const (
	tag = "database client"
)

type GormConfig struct {
	Dial       gorm.Dialector
	GormConfig gorm.Config
}

func InitGormDatabase(cf *GormConfig) (*gorm.DB, error) {
	db, err := gorm.Open(cf.Dial, &cf.GormConfig)
	if err != nil {
		return nil, err
	}
	err = db.Exec("SELECT 1").Error
	if err != nil {
		return nil, err
	}
	return db, nil
}

type BunConfig struct {
	Sql     *sql.DB
	Dialect schema.Dialect
	Opts    []bun.DBOption
}

func NewBunDatabase(cf *BunConfig) *bun.DB {
	db := bun.NewDB(cf.Sql, cf.Dialect, cf.Opts...)

	err := db.PingContext(context.Background())
	if err != nil {
		panic(err)
	}

	return db
}

type Config struct {
	Host         string  `env:"DB_HOST"`
	Port         int     `env:"DB_PORT"`
	Username     string  `env:"DB_USERNAME"`
	Password     string  `env:"DB_PASSWORD"`
	DatabaseName string  `env:"DB_NAME"`
	SSLMode      *string `env:"DB_SSL_MODE"`
	MaxOpenConns *int    `env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns *int    `env:"DB_MAX_IDLE_CONNS"`
}

func NewDatabase(ctx context.Context, cfg Config) *sql.DB {

	logger.Info(ctx, "connecting to database", slog.String("host", cfg.Host), slog.String("db-name", cfg.DatabaseName), logger.LogAttrTag(tag))

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DatabaseName,
		lo.FromPtrOr(cfg.SSLMode, "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error(ctx, "failed to open database connection", logger.LogAttrError(err), logger.LogAttrTag(tag))
		panic(err)
	}

	if err := db.PingContext(ctx); err != nil {
		logger.Error(ctx, "failed to ping database", logger.LogAttrError(err), logger.LogAttrTag(tag))
		panic(err)
	}

	db.SetMaxOpenConns(lo.FromPtrOr(cfg.MaxOpenConns, 25))
	db.SetMaxIdleConns(lo.FromPtrOr(cfg.MaxIdleConns, 5))

	logger.Info(ctx, "connect database success", slog.String("host", cfg.Host), slog.String("db-name", cfg.DatabaseName), logger.LogAttrTag(tag))
	return db
}
