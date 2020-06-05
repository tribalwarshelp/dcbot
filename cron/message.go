package cron

import (
	"fmt"

	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

type messageData struct {
	world            string
	date             string
	village          string
	oldOwnerName     string
	oldOwnerTribeTag string
	newOwnerName     string
	newOwnerTribeTag string
}

func newMessageData(world string, ennoblement *shared_models.Ennoblement) messageData {
	data := messageData{
		date:             formatDateOfConquest(utils.GetLocation(utils.LanguageCodeFromWorldName(world)), ennoblement.EnnobledAt),
		world:            world,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
	}
	if !isVillageNil(ennoblement.Village) {
		data.village = fmt.Sprintf("%s (%d|%d)", ennoblement.Village.Name, ennoblement.Village.X, ennoblement.Village.Y)
	}
	if !isPlayerNil(ennoblement.OldOwner) {
		data.oldOwnerName = ennoblement.OldOwner.Name
	}
	if !isPlayerTribeNil(ennoblement.OldOwner) {
		data.oldOwnerTribeTag = ennoblement.OldOwner.Tribe.Tag
	}
	if !isPlayerNil(ennoblement.NewOwner) {
		data.newOwnerName = ennoblement.NewOwner.Name
	}
	if !isPlayerTribeNil(ennoblement.NewOwner) {
		data.newOwnerTribeTag = ennoblement.NewOwner.Tribe.Tag
	}
	return data
}

func formatMsgAboutVillageLost(msgData messageData) string {
	return fmt.Sprintf(`**%s** %s: Wioska **%s** gracza **%s** (**%s**) została stracona na rzecz **%s** (**%s**)`,
		msgData.world,
		msgData.date,
		msgData.village,
		msgData.oldOwnerName,
		msgData.oldOwnerTribeTag,
		msgData.newOwnerName,
		msgData.newOwnerTribeTag)
}

func formatMsgAboutVillageConquest(msgData messageData) string {
	return fmt.Sprintf(`**%s** %s: Gracz **%s** (**%s**) podbił wioskę **%s** od gracza **%s** (**%s**)`,
		msgData.world,
		msgData.date,
		msgData.newOwnerName,
		msgData.newOwnerTribeTag,
		msgData.village,
		msgData.oldOwnerName,
		msgData.oldOwnerTribeTag)
}
