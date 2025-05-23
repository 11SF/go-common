package postgres

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Username string
	Password string
	Port     int
	DBName   string
	SSLMode  string
	TimeZone string
}

func ConnectPostgres(cf *Config) (gorm.Dialector, error) {
	if cf.TimeZone == "" {
		cf.TimeZone = "Asia/bangkok"
	}
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v", cf.Host, cf.Username, cf.Password, cf.DBName, cf.Port, cf.SSLMode, cf.TimeZone)
	dial := postgres.Open(dsn)
	return dial, nil
}

func NewPostgresConnection(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	return db
}
