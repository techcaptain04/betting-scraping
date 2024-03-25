package query

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConfig struct {
	Username string `mapstructure:"database.username"`
	Password string `mapstructure:"database.password"`
	DB       string `mapstructure:"database.db"`
	Host     string `mapstructure:"database.host"`
	Port     string `mapstructure:"database.port"`
}

type Category struct {
	ID    int32  `gorm:"primarykey"`
	Title string `gorm:"size:256"`
	Url   string `gorm:"size:256"`
}

var gdbContext *gorm.DB

func ConfigDatabase() {
	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/url?charset=utf8mb3&parseTime=True&loc=Local"
	gdbContext, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("DB connection error: ", err)
	}
	defer gdbContext.AutoMigrate(Category{})
}
