package utils

import (
	"github.com/tribalwarshelp/shared/models"
)

func IsPlayerNil(player *models.Player) bool {
	return player == nil
}

func IsPlayerTribeNil(player *models.Player) bool {
	return IsPlayerNil(player) || player.Tribe == nil
}

func IsVillageNil(village *models.Village) bool {
	return village == nil
}
