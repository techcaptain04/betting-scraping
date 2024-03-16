# Sportbook Scraper
 
A bot & web scraper to notify users when a discrepancy between a "dumb" and "smart" book is detected

## Technologies
- Golang
- MySQL (for caching game predictions)
- Docker & docker-compose
- https://github.com/gocolly/colly
- https://github.com/twilio/twilio-go

## Process Lifecycle
- App starts
- Creates a `query` service & channels
- Creates scrapers with the `query` channels
    - Refreshes the page every `x` seconds
    - Goes through each game
        - Checks cache
        - If the game doesn't exist, cache it & send to the `newGame` channel
- When the `newGame` channel receives a game:
    - Fetch all games with the same teams & date
    - Compare odds: if a "dumb" book has crazier odds than a "smart" book, or if the odds are crazy high for one game, send an SMS message

## Project Structure
```
C:.
│   go.mod
│   main.go
│   
├───cache (handles mysql operations)
│       cache.go
│       
├───query (compares game data to determine if a message should be sent)
│       query.go
│       
├───scrapers (colly scrapers for each website)
│   │   scraper.go (scraper interface)
│   │   
│   ├───dumb
│   │       bovada.go
│   │       ...
│   │
│   └───smart
│           betonline.go
│           ...
│
└───sms (for sending sms messages)
        sms.go
```

##  Scraper Interface
```go
package scraper

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
	// a ratio of odds for either team to win
	// based in terms of $100. if the odds are
	// <0, it means the amount of money you
	// must spend to earn $100. >0 is the amount
	// you could earn per $100 bet
	// if the team you bet on wins
	// index 0 is the odds of 1, index 1 is the odds of team 2
	Odds  []float64 `json:"odds"`
	Teams []string  `json:"teams"`
}
```

## "Dumb" books
- https://www.bovada.lv/
- https://www.mybookie.ag/
- https://www.purewage.com/

## "Smart" books
- https://mi.betrivers.com/?page=sportsbook&feed=featured#home
- https://sportsbook.fanduel.com/
- https://www.betonline.ag/