package utils

import (
	"github.com/tribalwarshelp/shared/models"
)

func LanguageCodeFromWorldName(world string) models.LanguageTag {
	if len(world) < 2 {
		return ""
	}
	return models.LanguageTag(world[0:2])
}
