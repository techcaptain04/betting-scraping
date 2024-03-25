module github.com/ferretcode-freelancing/sportsbook-scraper

go 1.21.6

replace service.com/dumb => ./scrapers/dumb

require service.com/dumb v0.0.0-00010101000000-000000000000

replace service.com/smart => ./scrapers/smart

require (
	gorm.io/driver/mysql v1.5.6
	gorm.io/gorm v1.25.8
)

require (
	github.com/PuerkitoBio/goquery v1.9.1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/antchfx/htmlquery v1.3.0 // indirect
	github.com/antchfx/xmlquery v1.3.18 // indirect
	github.com/antchfx/xpath v1.2.5 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocolly/colly v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/saintfish/chardet v0.0.0-20230101081208-5e3ef4b5456d // indirect
	github.com/temoto/robotstxt v1.1.2 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

require (
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/uuid v1.6.0
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/twilio/twilio-go v1.19.0
)

replace query.com/config => ./query
