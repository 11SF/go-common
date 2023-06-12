package database

import (
	"gorm.io/gorm"
)

type Config struct {
	Dial       gorm.Dialector
	GormConfig gorm.Config
}

func InitDatabase(cf *Config) (*gorm.DB, error) {
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
