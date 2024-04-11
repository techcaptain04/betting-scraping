package smart

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ferretcode-freelancing/sportsbook-scraper/cache"
	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/defaults"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type BetOnline struct {
	Name    string
	Scraper BetOnlineScraper
}

type BetOnlineScraper struct {
	Browser    *rod.Browser
	Categories []cache.CategoryURL
	DB         *gorm.DB
	URL        string
}

func RemoveDuplicate(url []cache.CategoryURL) []cache.CategoryURL {
	allKeys := make(map[string]bool)
	list := []cache.CategoryURL{}
	for _, item := range url {
		if _, value := allKeys[item.CategoryURL]; !value {
			allKeys[item.CategoryURL] = true
			list = append(list, item)
		}
	}
	return list
}

func NewBetOnline(db *gorm.DB) (BetOnline, error) {
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
		Where("base_url = ? ", "https://betonline.ag").
		Find(&categories).
		Error

	if err != nil {
		return BetOnline{}, err
	}

	return BetOnline{
		Name: "betonline",
		Scraper: BetOnlineScraper{
			Browser: browser,
			URL:     "https://sports.betonline.ag/sportsbook",
		},
	}, nil
}

func (b *BetOnlineScraper) GetURLs() ([]cache.CategoryURL, error) {
	var urls []cache.CategoryURL

	page, err := stealth.Page(b.Browser)
	if err != nil {
		return urls, err
	}

	page.MustSetViewport(1920, 1080, 1, false)

	err = page.Navigate("https://sports.betonline.ag/sportsbook")
	if err != nil {
		return urls, err
	}

	time.Sleep(10 * time.Second)

	dropdownSelector := "div.sidebarNavigation-dropDown"
	dropdownContentSelector := "div.sidebarNavigation-dropDownContentWrapper > div.sidebarNavigation-dropDownContent > div.sidebarNavigation-dropDown > div.sidebarNavigation-dropDownContentWrapper > div.sidebarNavigation-dropDownContent > div.sidebarNavigation-linkWrapper > a"

	dropdownButtons, err := page.Elements(dropdownSelector)
	if err != nil {
		return urls, err
	}

	for _, button := range dropdownButtons {
		err = button.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			return urls, err
		}

		content, err := page.Elements(dropdownSelector + " > " + dropdownContentSelector)
		if errors.Is(err, &rod.ElementNotFoundError{}) {
			continue
		}

		for _, link := range content {
			href, err := link.Attribute("href")
			if err != nil {
				return urls, err
			}

			urls = append(urls, cache.CategoryURL{
				BaseURL:     "https://betonline.ag",
				CategoryURL: *href,
			})
		}
	}

	return RemoveDuplicate(urls), err
}

func (b *BetOnlineScraper) GetProps(newProps chan scraper.Props, errChan chan error) {
	defaults.Show = true
	page, err := stealth.Page(b.Browser.NoDefaultDevice())
	HandleError(err, errChan)

	err = page.Navigate("https://sports.betonline.ag/sportsbook/props")
	HandleError(err, errChan)

	page.MustSetViewport(1920, 1080, 1, false)
	time.Sleep(30 * time.Second)

	f := page.MustElement("div.simplebar-wrapper div.simplebar-mask div.simplebar-content iframe").MustFrame()
	if f == nil {
		errChan <- errors.New("iframe nil or nonexistent")
		return
	}

	p := page.Browser().MustPageFromTargetID(proto.TargetTargetID(f.FrameID))

	mlbButtons := p.MustElements("div.ligues-slider__list.sports div div.ligues-slider__item.sportNames")

	for _, mlbButton := range mlbButtons {
		if mlbButton.MustElement("div.ligues-slider__ligue-name.cap").MustText() == "MLB" {
			err := mlbButton.Click(proto.InputMouseButtonLeft, 1)

			time.Sleep(10 * time.Second)

			HandleError(err, errChan)
			break
		}
	}

	typeButton := p.MustElement("div.main-content__wrapper div.main-markets div.main-markets__list div:nth-child(10)")
	if typeButton == nil {
		errChan <- errors.New("typebutton is nil")
		return
	}

	typeButton.Click(proto.InputMouseButtonLeft, 1)
	time.Sleep(10 * time.Second)

	// typeButtons := p.MustElements("div.main-content__wrapper div.main-markets div.main-markets__list div.main-markets__item")
	// fmt.Println("total buttons: ", len(typeButtons))
	// for _, button := range typeButtons {
	// 	text := button.MustElement("p.cap").MustText()
	// 	fmt.Println("text: ", text)
	// 	if text == "Total Bases (from Hits)" {
	// 		err := button.Click(proto.InputMouseButtonLeft, 1)
	// 		if err != nil {
	// 			fmt.Println("error: ", err)
	// 		}
	// 		break
	// 	}
	// }

	ticker := time.NewTicker(5 * time.Second)
	subTypeButtons := p.MustElements("div.main-content__wrapper div.main-stats div.main-stats__item.main-stat div.main-stat__header")

	for i, button := range subTypeButtons {
		if i == 0 {
			continue
		}
		button.Click(proto.InputMouseButtonLeft, 1)
		<-ticker.C
	}

	ticker = time.NewTicker(20 * time.Second)
	titleButtons := p.MustElements("div.main-stat__content div.tiered-block div.tiered-block__item div.tiered-block__item__top")

	for {
		<-ticker.C
		for _, button := range titleButtons {
			title := button.MustElement("div.tiered-block__title p.tiered-block__player-team")
			date := button.MustElement("div.tiered-block__title p.tiered-block__player-date.cap")

			time.Sleep(3 * time.Second)

			fmt.Println("title: ", title.MustText(), " date: ", date.MustText())

			items := p.MustElements("div.tiered-block__under-level-block--opened div.shots-block div.shots-block__player")

			for _, item := range items {
				player := item.MustElement("p.shots-block__player-name").MustText()
				values := item.MustElements("div.shots-block__player-values div.markets-slider__list.props div.markets-slider__item")

				amounts := []float64{}
				odds := []float64{}

				for _, value := range values {
					amount := value.MustElement("p.markets-slider__amount").MustText()
					stat := value.MustElement("p.markets-slider__stat").MustText()

					amountInt, err := strconv.ParseFloat(amount, 64)
					HandleError(err, errChan)
					oddsInt, err := strconv.ParseFloat(stat, 64)
					HandleError(err, errChan)

					amounts = append(amounts, amountInt)
					odds = append(odds, oddsInt)
				}

				err = b.DB.Model(&scraper.PropPlayer{}).Create(&scraper.PropPlayer{
					GameName: title.MustText(),
					Name:     player,
					Amounts:  amounts,
					Odds:     odds,
				}).Error
				HandleError(err, errChan)
			}

			teams := strings.Split(title.MustText(), " @ ")

			for i := range teams {
				teams[i] = strings.Trim(teams[i], " ")
			}

			newProps <- scraper.Props{
				Name:  title.MustText(),
				Date:  date.MustText(),
				Teams: teams,
			}
		}

		err = page.Reload()
		HandleError(err, errChan)
	}

}

func (b *BetOnlineScraper) GetGames(newGame chan scraper.Game, errChan chan error) {
	defer b.Browser.Close()

	page, err := stealth.Page(b.Browser)
	HandleError(err, errChan)

	err = page.Navigate("https://sports.betonline.ag/sportsbook")
	HandleError(err, errChan)

	//TODO: change to 30 in prod
	ticker := time.NewTicker(120 * time.Second)

	gamesSelector := "div.css-1mxf4v2 > a > div.card-sections"
	teamsSelector := "div.card-participants-section > div.card-score-participants > div.participants > div.participants-item > p"
	timeSelector := "div.card-participants-section > div.card-time > p"
	moneyLineSelector := "div.css-jgx3z6 > div:nth-child(2) > button > div > p"

	for {
		<-ticker.C
		for _, category := range b.Categories {
			err = page.Navigate(category.BaseURL + category.CategoryURL)
			HandleError(err, errChan)

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
					game.Odds = pq.Float64Array{team1Odds, team2Odds}
					game.Teams = pq.StringArray{teams[i].MustText(), teams[i+1].MustText()}

					if i >= len(times) {
						game.Date = times[(i/2)-1].MustText()
					} else {
						game.Date = times[i].MustText()
					}

					newGame <- game
				}
			}
		}

		err = page.Reload()
		HandleError(err, errChan)
	}
}

func CacheGames(games []scraper.Game) error {
	return nil
}
