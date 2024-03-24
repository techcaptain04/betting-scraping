package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/scrapers/smart"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	u := launcher.NewUserMode().
		Leakless(false).
		UserDataDir("tmp/t").
		Set("disable-default-apps").
		Set("no-first-run").
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	newGame := make(chan scraper.Game)
	errChan := make(chan error)

	scrapers := smart.GetScrapers(browser)

	go scrapers.BetOnline.Scraper.GetGames(newGame, errChan)

	go func() {
		for {
			select {
			case err := <-errChan:
				log.Fatal(err)
			case newGame := <-newGame:
				fmt.Println(newGame)
			}
		}
	}()

	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
}
