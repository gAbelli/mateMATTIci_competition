package main

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDB() {
	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")

	if len(DB_HOST) == 0 {
		DB_HOST = "localhost"
	}
	if len(DB_USER) == 0 {
		DB_USER = "root"
	}
	if len(DB_PASSWORD) == 0 {
		DB_PASSWORD = ""
	}
	if len(DB_NAME) == 0 {
		DB_NAME = "matemattici_competition"
	}
	if len(DB_PORT) == 0 {
		DB_PORT = "3306"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DB_USER,
		DB_PASSWORD,
		DB_HOST,
		DB_PORT,
		DB_NAME,
	)
	new_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	db = new_db
	db.AutoMigrate(&Submission{}, &Problem{}, &Competition{})
}
