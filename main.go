package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"service.com/dumb"
)

type DBConfig struct {
	Username string `mapstructure:"database.username"`
	Password string `mapstructure:"database.password"`
	DB       string `mapstructure:"database.db"`
	Host     string `mapstructure:"database.host"`
	Port     string `mapstructure:"database.port"`
}

var db *gorm.DB

func main() {

	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/scraping?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("DB connection error: ", err)
	}
	if err := db.AutoMigrate(&dumb.Category{}, &dumb.ScrapingData{}); err != nil {
		fmt.Println("AutoMigrate error: ", err)
		return
	}
	bookie := new(dumb.MyBookie)
	bookie.Run(db)
}
