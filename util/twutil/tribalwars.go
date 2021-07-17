package twutil

import (
	"github.com/tribalwarshelp/shared/tw/twmodel"
)

func IsPlayerNil(player *twmodel.Player) bool {
	return player == nil
}

func IsPlayerTribeNil(player *twmodel.Player) bool {
	return IsPlayerNil(player) || player.Tribe == nil
}

func IsVillageNil(village *twmodel.Village) bool {
	return village == nil
}
