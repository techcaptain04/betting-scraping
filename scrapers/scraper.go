package scraper

import "github.com/lib/pq"

type Scraper struct {
	Name    string           `json:"name"`
	URL     string           `json:"url"`
	Scraper ScraperInterface // will be omitted from cache
}

type ScraperInterface interface {
	GetGames(newGame chan Game, err chan error) // scrape all games from each sportsbook
	CacheGames([]Game) error                    // place all games in mysql
}

type Props struct {
	Name   string         `json:"name"`
	Date   string         `json:"date"`
	Teams  pq.StringArray `json:"teams" gorm:"type:text[]"`
	Source int            `json:"source"`
}

type PropPlayer struct {
	GameName string          `json:"game_name"`
	Name     string          `json:"name"`
	Amounts  pq.Float64Array `json:"amounts" gorm:"type:numeric[]"`
	Odds     pq.Float64Array `json:"odds" gorm:"type:numeric[]"`
}

type LegalProps struct {
	Name  string         `json:"name"`
	Date  string         `json:"date"`
	Teams pq.StringArray `json:"teams" gorm:"type:text[]"`
}

type LegalPlayer struct {
	GameName   string  `json:"game_name"`
	Name       string  `json:"name"`
	Determiner float64 `json:"determiner"`
	Over       float64 `json:"over"`
	Under      float64 `json:"under"`
}

const (
	BETONLINE = iota
	BETRIVERS
	FANDUEL
	ESPN
	CAESARS
)

type FatalError struct {
	Source int
	Error  error
}

func (f *FatalError) SetError(err error) FatalError {
	f.Error = err
	return *f
}

type Game struct {
	Id string `json:"id"` // not scraped
	// a ratio of odds for either team to win
	// based in terms of $100. if the odds are
	// <0, it means the amount of money you
	// must spend to earn $100. >0 is the amount
	// you could earn per $100 bet
	// if the team you bet on wins
	// index 0 is the odds of 1, index 1 is the odds of team 2
	League  string          `json:"league" gorm:"type:text"`
	Title   string          `json:"title" gorm:"type:text"`
	OddType string          `json:"oddType" gorm:"type:text"`
	Teams   pq.StringArray  `json:"team" gorm:"type:text[]"`
	Odds    pq.Float64Array `json:"odd" gorm:"type:numeric[]"`
	Date    string          `json:"date"`
}
