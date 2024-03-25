package cache

import (
	"errors"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/google/uuid"
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
