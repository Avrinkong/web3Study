package config

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() error {
	database, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	DB = database
	log.Println("Database connection established")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
