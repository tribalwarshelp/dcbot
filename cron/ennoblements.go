package cron

import shared_models "github.com/tribalwarshelp/shared/models"

type ennoblements []*shared_models.Ennoblement

func (e ennoblements) tribeLostVillages(tribeID int) ennoblements {
	filtered := ennoblements{}
	for _, ennoblement := range e {
		if (!isPlayerTribeNil(ennoblement.NewOwner) && ennoblement.NewOwner.Tribe.ID == tribeID) ||
			isPlayerTribeNil(ennoblement.OldOwner) ||
			ennoblement.OldOwner.Tribe.ID != tribeID {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}

func (e ennoblements) tribeConqueredVillages(tribeID int) ennoblements {
	filtered := ennoblements{}
	for _, ennoblement := range e {
		if isPlayerTribeNil(ennoblement.NewOwner) ||
			ennoblement.NewOwner.Tribe.ID != tribeID ||
			(!isPlayerTribeNil(ennoblement.OldOwner) && ennoblement.OldOwner.Tribe.ID == tribeID) {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}
