package main

import (
	// "fmt"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	// scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/cache"
	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/scrapers/smart"
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

	// newGame := make(chan scraper.Game)
	props := make(chan scraper.Props)
	errChan := make(chan error)

	scrapers := smart.GetScrapers(cache.DB)

	scrapers.BetOnline.Scraper.GetProps(props, errChan)
	// urls, err := scrapers.BetOnline.Scraper.GetURLs()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = cache.StoreURLs(urls)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	go func() {
		for {
			select {
			case err := <-errChan:
				log.Fatal(err)
			case props := <-props:
				fmt.Println(props)
			}
		}
	}()

	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
}
