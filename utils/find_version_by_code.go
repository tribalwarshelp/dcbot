package utils

import "github.com/tribalwarshelp/shared/tw/twmodel"

func FindVersionByCode(versions []*twmodel.Version, code twmodel.VersionCode) *twmodel.Version {
	var v *twmodel.Version
	for _, version := range versions {
		if version.Code == code {
			v = version
			break
		}
	}
	return v
}
