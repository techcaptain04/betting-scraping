package main

import (
	// "fmt"

	"log"
	"os"
	"os/signal"
	"syscall"

	// scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/cache"
	"github.com/ferretcode-freelancing/sportsbook-scraper/query"
	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/scrapers/smart"
	"github.com/ferretcode-freelancing/sportsbook-scraper/sms"
	"github.com/joho/godotenv"
)

func main() {
	if _, err := os.Stat("./.env"); err == nil {
		err = godotenv.Load("./.env")

		if err != nil {
			log.Fatal(err)
		}
	}

	cache, err := cache.NewCache()

	if err != nil {
		log.Fatal(err)
	}

	props := make(chan scraper.Props)
	errChan := make(chan error)
	fatalErr := make(chan scraper.FatalError)

	scrapers := smart.GetScrapers(cache.DB)

	queryService := query.QueryService{}

	smsService, err := sms.NewSMS(cache.DB)

	if err != nil {
		log.Fatal(err)
	}

	go queryService.ProcessProps(
		scrapers,
		props,
		errChan,
		fatalErr,
		smsService,
	)

	scrapers.StartScrapers(props, errChan, fatalErr)

	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
}
