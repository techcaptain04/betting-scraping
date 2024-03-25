package dumb

import (
	"fmt"
	"log"
	"os"
	"time"

	"strings"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MyBookieInterface interface {
	GetUrls(db *gorm.DB)
	Run(db *gorm.DB)
	GetData(db *gorm.DB)
}

type MyBookie struct {
	BaseUrl    string
	Bookie_inf MyBookieInterface
}

type Category struct {
	Uuid   string `gorm:"primarykey"`
	Domain string `gorm:"size:256"`
	Title  string `gorm:"size:256"`
	Link   string `gorm:"size:256"`
}

type ScrapingData struct {
	Uuid        string `gorm:"primarykey"`
	Title       string `gorm:"size:256"`
	HTeam       string `gorm:"size:256"`
	VTeam       string `gorm:"size:256"`
	HSpread     string `gorm:"size:32"`
	VSpread     string `gorm:"size:32"`
	StartDate   time.Time
	EndDate     time.Time
	CurrentDate time.Time
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

func (this *MyBookie) Run(db *gorm.DB) {
	this.BaseUrl = "https://www.mybookie.ag/sportsbook/"
	this.GetUrls(db)
	this.GetData(db)
}

func (this *MyBookie) GetUrls(db *gorm.DB) {
	fmt.Println(this.BaseUrl)

	c := colly.NewCollector()
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	data := []byte("Hello, this is log.\n")
	file, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error appending to file: ", err)
		return
	}
	c.OnHTML("div.left-menu", func(item *colly.HTMLElement) {
		item.ForEach("div.accordion div.left-menu-group div.sub-items-menu ul li.sub-items-menu__body__item a", func(_ int, subItem *colly.HTMLElement) {
			var item Category
			temp := strings.Replace(subItem.Text, "\n", "", -1)
			item.Title = strings.Replace(temp, "'", "", -1)
			item.Link = "https://www.mybookie.ag" + strings.Replace(subItem.Attr("href"), "'", "", -1)
			file.Write([]byte("sub item url: " + item.Link))
			file.Write([]byte("sub item title: " + item.Title))

			if item.Link != "" {
				var items []Category
				db.Where("link = ?", item.Link).Find(&items)
				if len(items) == 0 {
					id := uuid.NewString()
					query := fmt.Sprintf("INSERT INTO `categories` (`uuid`, `domain`, `title`, `link`) VALUES ('%s', '%s', '%s', '%s')", id, "MyBookie", item.Title, item.Link)
					err := db.Exec(query)
					if err != nil {
						fmt.Println("Error inserting data:", err)
						return
					}
					fmt.Println("Data has been written to the MySQL database")
				}
			}
		})
	})

	c.Visit(this.BaseUrl)
}

func (this *MyBookie) GetData(db *gorm.DB) {
	var categories []Category
	db.Find(&categories)
	fmt.Println("categories: ", len(categories))
	for _, category := range categories {
		var scData []ScrapingData
		c := colly.NewCollector()

		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL)
		})

		c.OnError(func(_ *colly.Response, err error) {
			log.Println("Something went wrong:", err)
		})

		c.OnResponse(func(r *colly.Response) {
			fmt.Println("Visited", r.Request.URL)
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
				tempEndDate := endDates.Eq(i).Text()
				tempStartDate := startDates.Eq(i).Text()
				endDateStr := GetTimeStr(tempEndDate)
				startDateStr := GetTimeStr(tempStartDate)
				endDate := ConvertTimeStamp(endDateStr)
				startDate := ConvertTimeStamp(startDateStr)
				currentDate := time.Now()
				scData = append(scData, ScrapingData{
					Title:       category.Title,
					HTeam:       homeTeam,
					VTeam:       visitorTeam,
					HSpread:     homeSpread,
					VSpread:     visitorSpread,
					StartDate:   startDate,
					EndDate:     endDate,
					CurrentDate: currentDate})
				id := uuid.NewString()
				var tempData []ScrapingData
				db.Where("start_date = ? AND h_team = ? AND v_team = ? AND title = ?", startDateStr, homeTeam, visitorTeam, category.Title).Find(&tempData)
				fmt.Println("start_date num: ", len(tempData))
				if len(tempData) == 0 {
					query := fmt.Sprintf("INSERT INTO `scraping_data` (`uuid`, `title`, `h_team`, `v_team`, `h_spread`, `v_spread`, `start_date`, `end_date`, `current_date`) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')",
						id, category.Title, homeTeam, visitorTeam, homeSpread, visitorSpread, startDate, endDate, currentDate)
					err := db.Exec(query)
					if err != nil {
						fmt.Println("Error inserting data:", err)
						return
					}
					fmt.Println("Data has been written to the MySQL database")
				}
			}
		})

		c.Visit(category.Link)
		for _, data := range scData {
			fmt.Println(
				"title: ", data.Title,
				"home team: ", data.HTeam,
				"visitor team: ", data.VTeam,
				"home spread: ", data.HSpread,
				"visitor spread: ", data.VSpread,
				"start date: ", data.StartDate,
				"end date: ", data.EndDate,
				"current date: ", data.CurrentDate)
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
