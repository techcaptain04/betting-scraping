package smart

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ferretcode-freelancing/sportsbook-scraper/cache"
	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type BetRivers struct {
	Name    string
	Scraper BetRiversScraper
}

type BetRiversScraper struct {
	Browser    *rod.Browser
	Categories []cache.CategoryURL
	DB         *gorm.DB
	URL        string
}

func NewBetRivers(db *gorm.DB) (BetRivers, error) {
	u := launcher.NewUserMode().
		Leakless(false).
		UserDataDir("tmp/t").
		Set("disable-default-apps").
		Set("no-first-run").
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	var categories []cache.CategoryURL

	err := db.
		Model(&cache.CategoryURL{}).
		Where("base_url = ? ", "https://mi.betrivers.com").
		Find(&categories).
		Error

	if err != nil {
		return BetRivers{}, err
	}

	return BetRivers{
		Name: "BetRivers",
		Scraper: BetRiversScraper{
			Browser: browser,
			URL:     "https://mi.betrivers.com/?page=sportsbook&feed=featured#home",
		},
	}, nil
}

func (b *BetRiversScraper) GetProps(
	newProps chan scraper.Props,
	errChan chan error,
	fatalError chan scraper.FatalError,
) {
	fatalErr := scraper.FatalError{
		Source: scraper.BETRIVERS,
	}
	defer b.Browser.Close()

	page, err := stealth.Page(b.Browser)
	HandleError(err, errChan)
	page.MustSetViewport(1920, 1080, 1, false)

	err = page.Navigate("https://mi.betrivers.com/?page=sportsbook&feed=featured#home")
	HandleError(err, errChan)
	time.Sleep(30 * time.Second)
	sportsButtons, err := page.Elements("#rsi-product-navigation-widget-bar-container > nav > div button")
	HandleError(err, errChan)
	// fmt.Println("buttons: ", len(sportsButtons))
	if len(sportsButtons) == 0 {
		fmt.Println("there is no button")
		return
	}
	for _, button := range sportsButtons {
		sportsName := button.MustAttribute("aria-label")
		if *sportsName == "Navigate to MLB" {
			// fmt.Println("Navigate to MLB")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			HandleError(err, errChan)
			time.Sleep(10 * time.Second)
			break
		}
	}
	propButtons := page.MustElements("div#sportsbook-page div#rsi-sports-feed > div > div > nav:nth-child(3) > div > div > button")
	// fmt.Println("propButtons", len(propButtons))
	if len(propButtons) == 0 {
		fmt.Println("no props")
		return
	}
	for _, button := range propButtons {
		prop := button.MustElement("span").MustText()
		if prop == "Player Props" {
			// fmt.Println("player props")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			HandleError(err, errChan)
			time.Sleep(10 * time.Second)
			break
		}
	}

	typeButtons, err := page.Elements("div#sportsbook-page div#rsi-sports-feed > div > div > div.sc-crwTFP.hutnOD div.sc-ivTmOn.bTibLZ div button")
	HandleError(err, errChan)
	// fmt.Println("typeButtons: ", len(typeButtons))
	if len(typeButtons) == 0 {
		fmt.Println("no types")
		return
	}
	for _, button := range typeButtons {
		if button.MustText() == "Total Bases" {
			// fmt.Println("Total Bases")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			HandleError(err, errChan)
			time.Sleep(10 * time.Second)
			break
		}
	}
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		items := page.MustElements("div#sportsbook-page div#rsi-sports-feed article.sc-ilxaJO.ldRamq")
		// fmt.Println("items: ", len(items))
		var team string
		var date string
		for _, item := range items {
			moreButton := item.MustElement("div.sc-eBOtmg.ezwdhJ button.sc-dAIixb.cRHXkq")
			moreButton.Click(proto.InputMouseButtonLeft, 1)
			<-ticker.C
			players := []string{}
			overs := []float64{}
			unders := []float64{}
			determiners := []float64{}

			temp, err := item.Element("div.sc-gNLcUQ.egUfuw")
			HandleError(err, errChan)
			titleData := temp.MustAttribute("aria-label")
			if titleData != nil {
				// fmt.Println("data: ", *titleData)
				str := *titleData
				results := strings.Split(str, ",")
				for i, result := range results {
					if i == 0 {
						team = result
						// fmt.Println("Title: ", result)
					} else if i == 1 {
						date = result
						// fmt.Println("Date: ", result)
					}
				}
			}

			data := temp.MustElement("div.sc-eBOtmg.ezwdhJ div.sc-ftNBoD.jUEglE")
			playersDOM := data.MustElements("div.sc-kRjaKC div.sc-jmjsKF div.sc-dRqsoR.sc-fWuLJ.jRUQXV.dzvaVD")
			oddsDOM := data.MustElements("div.sc-iaSSRK div.sc-iaSSRK")

			for _, item := range playersDOM {
				player := item.MustText()
				players = append(players, player)
				// fmt.Println("player: ", player)
			}

			for _, item := range oddsDOM {
				temps := item.MustElements("div.sc-fqEDVf button.sc-iKKmqK")
				for i, temp := range temps {
					amount := temp.MustElement("div.sc-dloDHE").MustText()
					stat := temp.MustElement("div.sc-hBYLEG ul.sc-imZoBe li").MustText()
					if i%2 == 0 {
						determiner, err := strconv.ParseFloat(strings.ReplaceAll(amount, "O\u00a0", ""), 64)
						determiners = append(determiners, determiner)
						HandleError(err, errChan)
						over, err := strconv.ParseFloat(stat, 64)
						HandleError(err, errChan)
						overs = append(overs, over)
					} else {
						HandleError(err, errChan)
						under, err := strconv.ParseFloat(stat, 64)
						unders = append(unders, under)
						HandleError(err, errChan)
					}
					// fmt.Println("stat: ", stat)
					// fmt.Println("amount: ", amount)
				}
			}
			title := *titleData
			for i := range players {
				err = b.DB.Model(&scraper.LegalPlayer{}).Create(&scraper.LegalPlayer{
					GameName:   title,
					Name:       players[i],
					Determiner: determiners[i],
					Over:       overs[i],
					Under:      unders[i],
				}).Error
				HandleError(err, errChan)
			}

			teams := strings.Split(team, " @ ")
			for i := range teams {
				teams[i] = strings.Trim(teams[i], " ")
			}

			newProps <- scraper.Props{
				Name:  title,
				Date:  date,
				Teams: teams,
			}
		}

		err = page.Reload()
		if err != nil {
			fatalError <- fatalErr.SetError(err)
			return
		}
	}
}

func (b *BetRiversScraper) GetGames(newGame chan scraper.Game, errChan chan error) {
	defer b.Browser.Close()

	page, err := stealth.Page(b.Browser)
	HandleError(err, errChan)

	page.MustSetViewport(1920, 1080, 1, false)

	err = page.Navigate(b.URL)
	HandleError(err, errChan)

	ticker := time.NewTicker(120 * time.Second)

	dropdownSelector := "ul.sc-doxBbc > li > div.sc-cszLSI > button.sc-lcWsae"
	dropdownContentSelector := `ul.sc-doxBbc > li > div[id*="sublist-items-panel"] > ul > li > div > button`

	navButtonSelector := "button.sc-fXnaaH:nth-child(1)"

	moneyLineSelector := "sc-jVUQqh > div > article > div > div > div.sc-hrRaah > div:nth-child(4) > div > div:nth-child(2) > div"
	teamSelector := "sc-jVUQqh > div > article > div > div > div.sc-hrRaah > div:nth-child(2) > div > div > div"

	for {
		<-ticker.C

		dropdowns, err := page.Elements(dropdownSelector)
		HandleError(err, errChan)

		for _, button := range dropdowns {
			err = button.Click(proto.InputMouseButtonLeft, 1)
			HandleError(err, errChan)

			subTypes, err := page.Elements(dropdownContentSelector)
			HandleError(err, errChan)

			for _, subType := range subTypes {
				err = subType.Click(proto.InputMouseButtonLeft, 1)
				HandleError(err, errChan)

				games, err := page.Element(navButtonSelector)
				HandleError(err, errChan)

				err = games.Click(proto.InputMouseButtonLeft, 1)
				HandleError(err, errChan)

				moneyLines, err := page.Elements(moneyLineSelector)
				HandleError(err, errChan)

				allTeams, err := page.Elements(teamSelector)
				HandleError(err, errChan)

				for i := range moneyLines {
					game := scraper.Game{}

					if i%2 == 0 && i != len(allTeams)-1 {
						team1Odds, err := strconv.ParseFloat(strings.Replace(moneyLines[i].MustText(), "+", "", -1), 64)
						HandleError(err, errChan)

						team2Odds, err := strconv.ParseFloat(strings.Replace(moneyLines[i+1].MustText(), "+", "", -1), 64)
						HandleError(err, errChan)

						game.Id = uuid.NewString()
						game.Odds = pq.Float64Array{team1Odds, team2Odds}
						game.Teams = pq.StringArray{allTeams[i].MustText(), allTeams[i+1].MustText()}
					}

					newGame <- game
				}
			}
		}

		err = page.Reload()
		HandleError(err, errChan)
	}
}
