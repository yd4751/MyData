package database

import (
	"fmt"

	"mediastation/config"
	"mediastation/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db

	err = DB.AutoMigrate(
		&model.User{},
		&model.PlayHistory{},
		&model.Media{},
		&model.MediaSeries{},
	)
	if err != nil {
		return err
	}

	return nil
}
