package cron

import (
	"time"

	shared_models "github.com/tribalwarshelp/shared/models"
)

func filterEnnoblements(ennoblements []*shared_models.Ennoblement, t time.Time) []*shared_models.Ennoblement {
	filtered := []*shared_models.Ennoblement{}
	for _, ennoblement := range ennoblements {
		if ennoblement.EnnobledAt.In(time.UTC).Before(t) {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}

func isPlayerNil(player *shared_models.Player) bool {
	return player == nil
}

func isPlayerTribeNil(player *shared_models.Player) bool {
	return isPlayerNil(player) || player.Tribe == nil
}

func isVillageNil(village *shared_models.Village) bool {
	return village == nil
}