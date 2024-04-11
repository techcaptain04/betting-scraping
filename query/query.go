package query

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"

	scraper "github.com/ferretcode-freelancing/sportsbook-scraper/scrapers"
	"github.com/ferretcode-freelancing/sportsbook-scraper/scrapers/smart"
	"github.com/ferretcode-freelancing/sportsbook-scraper/sms"
	"gorm.io/gorm"
)

type QueryService struct {
	DB *gorm.DB
}

func (q *QueryService) ProcessProps(
	scrapers smart.Scrapers,
	newProps chan scraper.Props,
	errChan chan error,
	fatalError chan scraper.FatalError,
	sms sms.SMS,
) {
	for {
		select {
		case props := <-newProps:
			var propsPlayers []scraper.PropPlayer

			err := q.DB.
				Model(&scraper.PropPlayer{}).
				Where("game_name = ?", props.Name).
				Find(&propsPlayers).
				Error

			if err != nil {
				propError(props.Name, err)

				continue
			}

			var legalPlayers []scraper.LegalPlayer

			err = q.DB.
				Model(&scraper.LegalPlayer{}).
				Where("game_name = ?", props.Name).
				Find(&legalPlayers).
				Error

			if err != nil {
				propError(props.Name, err)

				continue
			}

			for _, legalPlayer := range legalPlayers {
				propsPlayer := slices.IndexFunc(propsPlayers, func(pp scraper.PropPlayer) bool {
					return pp.Name == legalPlayer.Name
				})

				for i, amount := range propsPlayers[propsPlayer].Amounts {
					propOdds := propsPlayers[propsPlayer].Odds[i]

					mappedOdds := []string{}

					if propOdds > 0 {
						mappedOdds = append(mappedOdds, fmt.Sprintf("+%f", propOdds))
					} else {
						mappedOdds = append(mappedOdds, fmt.Sprintf("%f", propOdds))
					}

					if amount >= legalPlayer.Over {
						mappedOdds = append(mappedOdds, fmt.Sprintf("+%f", legalPlayer.Over))

						diff :=
							math.Abs(legalPlayer.Over-propOdds) / ((legalPlayer.Over + propOdds) / 2) * 100

						threshold := os.Getenv("COMPARISON_THRESHOLD")

						thresholdFloat, _ := strconv.ParseFloat(threshold, 64)

						if diff > thresholdFloat {
							sms.SendSMS(props.Source, props.Name, legalPlayer.Name, mappedOdds)
						}
					} else {
						mappedOdds = append(mappedOdds, fmt.Sprintf("%f", legalPlayer.Over))

						diff :=
							math.Abs(legalPlayer.Under-propOdds) / ((legalPlayer.Under + propOdds) / 2) * 100

						threshold := os.Getenv("COMPARISON_THRESHOLD")

						thresholdFloat, _ := strconv.ParseFloat(threshold, 64)

						if diff > thresholdFloat {
							sms.SendSMS(props.Source, props.Name, legalPlayer.Name, mappedOdds)
						}
					}
				}
			}
		case err := <-errChan:
			fmt.Println(err)
		case err := <-fatalError:
			fmt.Printf("New fatal error from source %d.\nError: %s\n", err.Source, err.Error.Error())
			switch err.Source {
			case scraper.BETONLINE:
				go scrapers.BetOnline.Scraper.GetProps(newProps, errChan, fatalError) // rerun if fail
			}
		}
	}
}

func propError(name string, err error) {
	fmt.Printf("Dropped %s with error: %s\n", name, err.Error())
}
