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

type Game struct {
	Id string `json:"id"` // not scraped
	// a ratio of odds for either team to win
	// based in terms of $100. if the odds are
	// <0, it means the amount of money you
	// must spend to earn $100. >0 is the amount
	// you could earn per $100 bet
	// if the team you bet on wins
	// index 0 is the odds of 1, index 1 is the odds of team 2
	Odds  pq.Float64Array `json:"odds" gorm:"type:numeric[]"`
	Teams pq.StringArray  `json:"teams" gorm:"type:text[]"`
	Date  string            `json:"date"`
}
