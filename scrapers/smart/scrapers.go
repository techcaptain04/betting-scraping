package smart

import "github.com/go-rod/rod"

type Scrapers struct {
	BetOnline BetOnline
}

func GetScrapers(rod *rod.Browser) Scrapers {
	return Scrapers{
		BetOnline: NewBetOnline(rod),
	}
}

func HandleError(err error, errChan chan error) {
	if err != nil {
		errChan <- err
	}
}
