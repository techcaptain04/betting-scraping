package dumb

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"strings"

	"github.com/ferretcode-freelancing/sportsbook-scraper/cache"
	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MyBookie struct {
	Name    string
	Scraper MyBookieScraper
}

type MyBookieScraper struct {
	URL        string
	Categories []cache.CategoryURL
}

type Date struct {
	Year   int32
	Month  int32
	Day    int32
	Hour   int32
	Minute int32
	Sec    int32
	Msec   int32
}

func NewMyBookie(db *gorm.DB) (MyBookie, error) {
	var categories []cache.CategoryURL
	baseUrl := "https://mybookie.ag/sportsbook"
	err := db.Model(&cache.CategoryURL{}).Where("base_url = ?", baseUrl).Find(&categories).Error

	if err != nil {
		return MyBookie{}, err
	}

	return MyBookie{
		Name: "MyBookie",
		Scraper: MyBookieScraper{
			URL:        baseUrl,
			Categories: categories,
		},
	}, nil
}

func (b *MyBookie) GetUrls() ([]cache.CategoryURL, error) {
	var urls []cache.CategoryURL

	c := colly.NewCollector()

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnHTML("div.left-menu", func(item *colly.HTMLElement) {
		item.ForEach("div.accordion div.left-menu-group div.sub-items-menu ul li.sub-items-menu__body__item a", func(_ int, subItem *colly.HTMLElement) {
			urls = append(urls, cache.CategoryURL{
				BaseURL:     "https://www.mybookie.ag",
				CategoryURL: strings.Replace(subItem.Attr("href"), "'", "", -1),
			})
		})
	})

	c.Visit(b.Scraper.URL)

	return urls, nil
}

func (b *MyBookie) GetData(newGame chan scraper.Game, errChan chan error) {
	for _, category := range b.Scraper.Categories {
		c := colly.NewCollector()

		ticker := time.NewTicker(10 * time.Second)

		for {
			<-ticker.C

			c.OnError(func(_ *colly.Response, err error) {
				log.Println("Something went wrong:", err)
				if err != nil {
					errChan <- err
				}
			})

			c.OnHTML("div.game-lines div.game-line", func(item *colly.HTMLElement) {
				homeTeams := item.DOM.Find("div.game-line__home-team a")
				visitorTeams := item.DOM.Find("div.game-line__visitor-team a")
				homeSpreads := item.DOM.Find("div.game-line__home-line > :first-child span")
				visitorSpreads := item.DOM.Find("div.game-line__visitor-line > :first-child span")
				startDates := item.DOM.Find("div.d-none > :first-child")
				endDates := item.DOM.Find("div.d-none > :nth-child(2)")

				numTeams := min(homeTeams.Length(), visitorTeams.Length(), homeSpreads.Length(), startDates.Length(), endDates.Length())

				for i := 0; i < numTeams; i++ {
					homeTeam := strings.Replace(homeTeams.Eq(i).Text(), "'", "", -1)
					visitorTeam := strings.Replace(visitorTeams.Eq(i).Text(), "'", "", -1)
					homeSpread := homeSpreads.Eq(i).Text()
					visitorSpread := visitorSpreads.Eq(i).Text()
					tempStartDate := startDates.Eq(i).Text()
					startDateStr := GetTimeStr(tempStartDate)
					startDate := ConvertTimeStamp(startDateStr)

					team1Odds, err := strconv.ParseFloat(strings.Replace(homeSpread, "+", "", -1), 64)
					if err != nil {
						errChan <- err
						return
					}

					team2Odds, err := strconv.ParseFloat(strings.Replace(visitorSpread, "+", "", -1), 64)
					if err != nil {
						errChan <- err
						return
					}

					game := scraper.Game{
						Id:   uuid.NewString(),
						Team: pq.StringArray{homeTeam, visitorTeam},
						Odd:  pq.Float64Array{team1Odds, team2Odds},
						Date: strconv.FormatInt(startDate.Unix(), 10),
					}

					newGame <- game
				}
			})

			c.Visit(category.BaseURL + category.CategoryURL)
		}

	}
}

func GetTimeStr(timeStr string) (converted string) {
	var date Date
	fmt.Sscanf(timeStr, "%04d - %02d - %02d %02d:%02d+%02d:%02d",
		&date.Year,
		&date.Month,
		&date.Day,
		&date.Hour,
		&date.Minute,
		&date.Sec,
		&date.Msec)

	converted = fmt.Sprintf("%04d-%02d-%02d %02d:00:00.000", date.Year,
		date.Month,
		date.Day,
		date.Hour)
	return converted
}

func ConvertTimeStamp(timeStr string) (timestamp time.Time) {
	timestamp, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return
	}
	fmt.Println("Time String", timestamp)
	return timestamp
}
