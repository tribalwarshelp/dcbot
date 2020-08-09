package cron

import (
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

type ennoblements []*shared_models.LiveEnnoblement

func (e ennoblements) getLastEnnoblement() *shared_models.LiveEnnoblement {
	length := len(e)
	if length <= 0 {
		return nil
	}
	return e[length-1]
}

func (e ennoblements) getLostVillagesByTribe(tribeID int) ennoblements {
	filtered := ennoblements{}
	for _, ennoblement := range e {
		if (!utils.IsPlayerTribeNil(ennoblement.NewOwner) && ennoblement.NewOwner.Tribe.ID == tribeID) ||
			utils.IsPlayerTribeNil(ennoblement.OldOwner) ||
			ennoblement.OldOwner.Tribe.ID != tribeID {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}

func (e ennoblements) getConqueredVillagesByTribe(tribeID int, showInternals bool) ennoblements {
	filtered := ennoblements{}
	for _, ennoblement := range e {
		if utils.IsPlayerTribeNil(ennoblement.NewOwner) ||
			ennoblement.NewOwner.Tribe.ID != tribeID ||
			(!showInternals && !utils.IsPlayerTribeNil(ennoblement.OldOwner) && ennoblement.OldOwner.Tribe.ID == tribeID) {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}
