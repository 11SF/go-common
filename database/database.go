package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"gorm.io/gorm"
)

type Config struct {
	Dial       gorm.Dialector
	GormConfig gorm.Config
}

func InitGormDatabase(cf *Config) (*gorm.DB, error) {
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
