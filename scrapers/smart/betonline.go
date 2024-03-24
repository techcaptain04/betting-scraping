package smart

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
	"github.com/google/uuid"
)

type BetOnline struct {
	Name    string
	Scraper BetOnlineScraper
}

type BetOnlineScraper struct {
	Browser *rod.Browser
	URL     string
}

func NewBetOnline(rod *rod.Browser) BetOnline {
	return BetOnline{
		Name: "betonline",
		Scraper: BetOnlineScraper{
			Browser: rod,
			URL:     "https://sports.betonline.ag/sportsbook",
		},
	}
}

func (b *BetOnlineScraper) GetGames(newGame chan scraper.Game, errChan chan error) {
	page, err := stealth.Page(b.Browser)
	HandleError(err, errChan)

	err = page.Navigate("https://sports.betonline.ag/sportsbook")
	HandleError(err, errChan)

	//TODO: change to 30 in prod
	ticker := time.NewTicker(10 * time.Second)

	gamesSelector := "div.css-1mxf4v2 > a > div.card-sections"
	teamsSelector := "div.card-participants-section > div.card-score-participants > div.participants > div.participants-item > p"
	timeSelector := "div.card-participants-section > div.card-time > p"
	moneyLineSelector := "div.css-jgx3z6 > div:nth-child(2) > button > div > p"

	for {
		<-ticker.C
		teams, err := page.Elements(gamesSelector + " > " + teamsSelector)
		HandleError(err, errChan)

		times, err := page.Elements(gamesSelector + " > " + timeSelector)
		HandleError(err, errChan)

		moneyLine, err := page.Elements(gamesSelector + " > " + moneyLineSelector)
		HandleError(err, errChan)

		fmt.Printf("num teams: %d\n", len(teams))
		fmt.Printf("num times: %d\n", len(times))
		fmt.Printf("num moneylines: %d\n", len(moneyLine))

		for i := range teams {
			game := scraper.Game{}

			if i%2 == 0 && i != len(teams)-1 {
				team1Odds, err := strconv.ParseFloat(strings.Replace(moneyLine[i].MustText(), "+", "", -1), 64)
				HandleError(err, errChan)

				team2Odds, err := strconv.ParseFloat(strings.Replace(moneyLine[i+1].MustText(), "+", "", -1), 64)
				HandleError(err, errChan)

				game.Id = uuid.NewString()
				game.Odds = []float64{team1Odds, team2Odds}
				game.Teams = []string{teams[i].MustText(), teams[i+1].MustText()}

				if i >= len(times) {
					game.Date = times[(i/2)-1].MustText()
				} else {
					game.Date = times[i].MustText()
				}

				newGame <- game
			}
		}

		err = page.Reload()
		HandleError(err, errChan)
	}
}

func CacheGames(games []scraper.Game) error {
	return nil
}
