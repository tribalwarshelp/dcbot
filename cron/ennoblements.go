package cron

import (
	"github.com/tribalwarshelp/shared/tw/twmodel"

	"github.com/tribalwarshelp/dcbot/util/twutil"
)

type ennoblements []*twmodel.Ennoblement

func (e ennoblements) getLastEnnoblement() *twmodel.Ennoblement {
	length := len(e)
	if length <= 0 {
		return nil
	}
	return e[length-1]
}

func (e ennoblements) getLostVillagesByTribe(tribeID int) ennoblements {
	filtered := ennoblements{}
	for _, ennoblement := range e {
		if (!twutil.IsPlayerTribeNil(ennoblement.NewOwner) && ennoblement.NewOwner.Tribe.ID == tribeID) ||
			twutil.IsPlayerTribeNil(ennoblement.OldOwner) ||
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
		if twutil.IsPlayerTribeNil(ennoblement.NewOwner) ||
			ennoblement.NewOwner.Tribe.ID != tribeID ||
			(!showInternals && !twutil.IsPlayerTribeNil(ennoblement.OldOwner) && ennoblement.OldOwner.Tribe.ID == tribeID) {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}
