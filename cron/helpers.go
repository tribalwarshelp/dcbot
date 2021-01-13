package cron

import (
	"time"

	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

func getRecentEnnoblements(ennoblements []*shared_models.LiveEnnoblement, t time.Time) []*shared_models.LiveEnnoblement {
	filtered := []*shared_models.LiveEnnoblement{}
	for _, ennoblement := range ennoblements {
		if ennoblement.EnnobledAt.Before(t) || ennoblement.EnnobledAt.Equal(t) {
			continue
		}
		filtered = append(filtered, ennoblement)
	}
	return filtered
}

func isBarbarian(p *shared_models.Player) bool {
	return utils.IsPlayerNil(p) || p.ID == 0
}
