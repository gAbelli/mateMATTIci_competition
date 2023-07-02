package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func SetupDB() {
	dsn := "root@tcp(127.0.0.1:3306)/matemattici_competition?charset=utf8mb4&parseTime=True&loc=Local"
	new_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	db = new_db
	db.AutoMigrate(&Submission{}, &Problem{}, &Competition{})
}
