package database

import "gorm.io/gorm"

// type Config struct {
// 	Host     string
// 	Password string
// 	Port     int
// 	DBName   string
// 	TimeZone string
// }

type Config struct {
	Dial       gorm.Dialector
	GormConfig gorm.Config
}

func (cf *Config) InitDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(cf.Dial, &cf.GormConfig)
	if err != nil {
		return nil, err
	}
	return db, nil
}
