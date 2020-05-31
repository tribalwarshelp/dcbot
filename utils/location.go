package utils

import "time"

func GetLocation(lang string) *time.Location {
	switch lang {
	case "pl":
		loc, _ := time.LoadLocation("Europe/Warsaw")
		return loc
	default:
		return time.UTC
	}
}
