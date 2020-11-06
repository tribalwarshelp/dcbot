package utils

import "github.com/tribalwarshelp/shared/models"

func FindVersionByCode(versions []*models.Version, code models.VersionCode) *models.Version {
	v := &models.Version{}
	for _, version := range versions {
		if version.Code == code {
			v = version
			break
		}
	}
	return v
}
