package cron

import (
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

func isBarbarian(p *shared_models.Player) bool {
	return utils.IsPlayerNil(p) || p.ID == 0
}
