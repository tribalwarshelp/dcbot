package cron

import (
	"time"

	shared_models "github.com/tribalwarshelp/shared/models"
)

func filterEnnoblements(ennoblements []*shared_models.LiveEnnoblement, t time.Time) []*shared_models.LiveEnnoblement {
	filtered := []*shared_models.LiveEnnoblement{}
	for _, ennoblement := range ennoblements {
		if ennoblement.EnnobledAt.Before(t) || ennoblement.EnnobledAt.Equal(t) {
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

func formatDateOfConquest(t time.Time) string {
	return t.Format(time.RFC3339)
}

func getLocation(timezone string) *time.Location {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
