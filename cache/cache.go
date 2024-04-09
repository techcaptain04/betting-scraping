package cache

import (
	"errors"
	"os"
	"strings"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/google/uuid"
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
	dsn := os.Getenv("DSN")
	if isLocal == "on" {
		dsn = os.Getenv("LDSN")
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
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
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

	err = DB.AutoMigrate(&scraper.Game{}, CategoryURL{})

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

func (c *Cache) WriteCache(game scraper.Game) error {
	game.Id = uuid.NewString()

	err := c.DB.Create(game).Error

	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetCache(date string, teams []string) (scraper.Game, error) {
	var game scraper.Game

	err := c.DB.Where("teams = ? AND date = ?", teams, date).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return scraper.Game{}, ErrGameNotFound
		}

		return scraper.Game{}, err
	}

	return game, nil
}
