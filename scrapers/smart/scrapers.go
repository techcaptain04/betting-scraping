package smart

import (
	"log"

	"gorm.io/gorm"
)

type Scrapers struct {
	BetOnline BetOnline
	BetRivers BetRivers
}

func GetScrapers(db *gorm.DB) Scrapers {
	betonline, err := NewBetOnline(db)
	if err != nil {
		log.Fatal(err)
	}
	betrivers, err := NewBetRivers(db)
	if err != nil {
		log.Fatal(err)
	}

	return Scrapers{
		BetOnline: betonline,
		BetRivers: betrivers,
	}
}

func HandleError(err error, errChan chan error) {
	if err != nil {
		errChan <- err
	}
}
