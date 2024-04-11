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

func (b *BetRiversScraper) GetProps() {
	defer b.Browser.Close()

	page, err := stealth.Page(b.Browser)
	if err != nil {
		fmt.Println("browser issue")
		return
	}
	page.MustSetViewport(1920, 1080, 1, false)

	err = page.Navigate("https://mi.betrivers.com/?page=sportsbook&feed=featured#home")
	if err != nil {
		return
	}
	time.Sleep(30 * time.Second)
	sportsButtons, err := page.Elements("#rsi-product-navigation-widget-bar-container > nav > div button")

	if err != nil {
		fmt.Println("mlb button error: ", err)
		return
	}
	fmt.Println("buttons: ", len(sportsButtons))
	if len(sportsButtons) == 0 {
		fmt.Println("there is no button")
		return
	}
	for _, button := range sportsButtons {
		sportsName := button.MustAttribute("aria-label")
		if *sportsName == "Navigate to MLB" {
			fmt.Println("Navigate to MLB")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			if err != nil {
				fmt.Println("sports button error: ", err)
			}
			time.Sleep(10 * time.Second)
			break
		}
	}
	propButtons := page.MustElements("div#sportsbook-page div#rsi-sports-feed > div > div > nav:nth-child(3) > div > div > button")
	fmt.Println("propButtons", len(propButtons))
	if len(propButtons) == 0 {
		fmt.Println("no props")
		return
	}
	for _, button := range propButtons {
		prop := button.MustElement("span").MustText()
		if prop == "Player Props" {
			fmt.Println("player props")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			if err != nil {
				fmt.Println("prop button error: ", err)
				return
			}
			time.Sleep(10 * time.Second)
			break
		}
	}

	typeButtons, err := page.Elements("div#sportsbook-page div#rsi-sports-feed > div > div > div.sc-crwTFP.hutnOD div.sc-ivTmOn.bTibLZ div button")
	if err != nil {
		fmt.Println("type buttons error: ", err)
		return
	}
	fmt.Println("typeButtons: ", len(typeButtons))
	if len(typeButtons) == 0 {
		fmt.Println("no types")
		return
	}
	for _, button := range typeButtons {
		if button.MustText() == "Total Bases" {
			fmt.Println("Total Bases")
			err := button.Click(proto.InputMouseButtonLeft, 1)
			if err != nil {
				fmt.Println("button click error: ", err)
				return
			}
			time.Sleep(10 * time.Second)
			break
		}
	}

	ticker := time.NewTicker(2 * time.Second)
	items := page.MustElements("div#sportsbook-page div#rsi-sports-feed article.sc-ilxaJO.ldRamq")
	fmt.Println("items: ", len(items))
	for _, item := range items {
		moreButton := item.MustElement("div.sc-eBOtmg.ezwdhJ button.sc-dAIixb.cRHXkq")
		moreButton.Click(proto.InputMouseButtonLeft, 1)
		<-ticker.C

		temp, err := item.Element("div.sc-gNLcUQ.egUfuw")
		if err != nil {
			fmt.Println("data errors: ", err)
		}
		titleData := temp.MustAttribute("aria-label")
		if titleData != nil {
			fmt.Println("data: ", *titleData)
			str := *titleData
			results := strings.Split(str, ",")
			for i, result := range results {
				if i == 0 {
					fmt.Println("Title: ", result)
				} else if i == 1 {
					fmt.Println("Date: ", result)
				}
			}
		}
		data := temp.MustElement("div.sc-eBOtmg.ezwdhJ div.sc-ftNBoD.jUEglE")
		players := data.MustElements("div.sc-kRjaKC div.sc-jmjsKF div.sc-dRqsoR.sc-fWuLJ.jRUQXV.dzvaVD")
		odds := data.MustElements("div.sc-iaSSRK div.sc-iaSSRK")
		for _, item := range players {
			player := item.MustText()
			fmt.Println("player: ", player)
		}
		for _, item := range odds {
			temps := item.MustElements("div.sc-fqEDVf button.sc-iKKmqK")
			for _, temp := range temps {
				stat := temp.MustElement("div.sc-dloDHE").MustText()
				fmt.Println("stat: ", stat)
				amount := temp.MustElement("div.sc-hBYLEG ul.sc-imZoBe li").MustText()
				fmt.Println("amount: ", amount)
			}
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
