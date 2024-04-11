package cache

import (
	"errors"
	"os"
	"strings"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/sms"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Cache struct {
	DB *gorm.DB
}

type CategoryURL struct {
	BaseURL     string
	CategoryURL string
}

var ErrGameNotFound = errors.New("the game was not found in the database")

func NewCache() (Cache, error) {
	isLocal := os.Getenv("LOCAL")
	dbChoice := os.Getenv("DB_CHOICE")

	dsn := os.Getenv("DSN")

	if isLocal == "on" {
		switch dbChoice {
		case "postgres":
			dsn = os.Getenv("LDSN_POSTGRES")
		case "mysql":
			dsn = os.Getenv("LDSN_MYSQL")
		}
	}

	if dsn == "" {
		return Cache{}, errors.New("the dsn env variable was not found")
	}

	dsn = strings.ReplaceAll(dsn, "\n", "")
	var DB *gorm.DB
	var err error
	if isLocal != "on" {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	} else {
		switch dbChoice {
		case "postgres":
			DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
			})
		case "mysql":
			DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
			})
		}

	}

	if err != nil {
		return Cache{}, err
	}

	sqlDB, err := DB.DB()

	if err != nil {
		return Cache{}, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	err = DB.AutoMigrate(&scraper.Game{}, &CategoryURL{}, &sms.User{})

	if err != nil {
		return Cache{}, err
	}

	return Cache{
		DB: DB,
	}, nil
}

func (c *Cache) StoreURLs(categories []CategoryURL) error {
	err := c.DB.Create(categories).Error

	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) WriteCache(props scraper.Props) error {
	err := c.DB.Create(props).Error

	if err != nil {
		return err
	}

	return nil
}
