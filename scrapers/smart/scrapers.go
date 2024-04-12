package smart

import (
	"log"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
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

func (s *Scrapers) StartScrapers(
	newProps chan scraper.Props,
	errChan chan error,
	fatalError chan scraper.FatalError,
) {
	go s.BetRivers.Scraper.GetProps(newProps, errChan, fatalError)
}

func HandleError(err error, errChan chan error) {
	if err != nil {
		errChan <- err
	}
}
