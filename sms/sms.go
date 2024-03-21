package sms

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"gorm.io/gorm"
)

type User struct {
	PhoneNumber string
}

type SMS struct {
	PhoneNumber  string
	TwilioClient twilio.RestClient
	Logger       slog.Logger
	Recipients   []string
}

func NewSMS(db *gorm.DB) (SMS, error) {
	sms := SMS{
		PhoneNumber:  os.Getenv("TWILIO_PHONE_NUMBER"),
		TwilioClient: *twilio.NewRestClient(),
		Logger:       *slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}

	recipients, err := sms.GetRecipients(db)

	if err != nil {
		return sms, err
	}

	sms.Recipients = recipients

	return sms, nil
}

func (s *SMS) SendSMS(siteURL string, teams []string, odds []float64) error {
	params := &api.CreateMessageParams{}

	params.SetBody(
		fmt.Sprintf(
			"New discrepancy detected on %s. The odds are %f - %f for teams %s and %s.",
			siteURL,
			odds[0],
			odds[1],
			teams[0],
			teams[1],
		),
	)
	params.SetFrom(s.PhoneNumber)

	for _, r := range s.Recipients {
		params.SetTo(r)

		res, err := s.TwilioClient.Api.CreateMessage(params)

		if err != nil {
			return err
		}

		s.Logger.Info("SMS message sent to", "number", r, "sid", res.Sid)
	}

	return nil
}

func (s *SMS) GetRecipients(db *gorm.DB) ([]string, error) {
	var users []User

	err := db.Model(&User{}).Find(&users).Error

	if err != nil {
		return []string{}, err
	}

	return []string{}, nil
}
