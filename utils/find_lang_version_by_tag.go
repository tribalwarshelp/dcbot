package utils

import "github.com/tribalwarshelp/shared/models"

func FindLangVersionByTag(langVersions []*models.LangVersion, tag models.LanguageTag) *models.LangVersion {
	lv := &models.LangVersion{}
	for _, langVersion := range langVersions {
		if langVersion.Tag == tag {
			lv = langVersion
			break
		}
	}
	return lv
}
