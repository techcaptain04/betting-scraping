package smart

type Scrapers struct {
	BetOnline BetOnline
}

func GetScrapers() Scrapers {
	return Scrapers{
		BetOnline: NewBetOnline(),
	}
}

func HandleError(err error, errChan chan error) {
	if err != nil {
		errChan <- err
	}
}
