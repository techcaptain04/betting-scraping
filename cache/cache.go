package cache

import (
	"errors"
	"os"
	"strings"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Cache struct {
	db *gorm.DB
}

type CategoryURL struct {
	BaseURL     string
	CategoryURL string
}

var ErrGameNotFound = errors.New("the game was not found in the database")

func NewCache() (Cache, error) {
	dsn := os.Getenv("DSN")

	if dsn == "" {
		return Cache{}, errors.New("the dsn env variable was not found")
	}

	dsn = strings.ReplaceAll(dsn, "\n", "")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return Cache{}, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return Cache{}, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	err = db.AutoMigrate(&scraper.Game{}, CategoryURL{})

	if err != nil {
		return Cache{}, err
	}

	return Cache{
		db: db,
	}, nil
}

func (c *Cache) StoreURLs(categories []CategoryURL) error {
	err := c.db.Create(categories).Error

	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) WriteCache(game scraper.Game) error {
	game.Id = uuid.NewString()

	err := c.db.Create(game).Error

	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetCache(date string, teams []string) (scraper.Game, error) {
	var game scraper.Game

	err := c.db.Where("teams = ? AND date = ?", teams, date).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return scraper.Game{}, ErrGameNotFound
		}

		return scraper.Game{}, err
	}

	return game, nil
}
