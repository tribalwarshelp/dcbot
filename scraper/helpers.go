package scraper

import "time"

var (
	TwstatsURLs = map[string]string{
		"pl": "https://pl.twstats.com",
	}
	Locations = loadLocations()
)

func loadLocations() map[string]*time.Location {
	m := make(map[string]*time.Location)
	m["pl"], _ = time.LoadLocation("Europe/Warsaw")
	return m
}
