package smart

import (
	"log"

	"gorm.io/gorm"
)

type Scrapers struct {
	BetOnline BetOnline
}

func GetScrapers(db *gorm.DB) Scrapers {
	betonline, err := NewBetOnline(db)

	if err != nil {
		log.Fatal(err)
	}

	return Scrapers{
		BetOnline: betonline,
	}
}

func HandleError(err error, errChan chan error) {
	if err != nil {
		errChan <- err
	}
}
