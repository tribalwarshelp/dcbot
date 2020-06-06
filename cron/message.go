package cron

import (
	"fmt"

	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

const (
	endpointTribeProfile   = "/game.php?screen=info_ally&id=%d"
	endpointPlayerProfile  = "/game.php?screen=info_player&id=%d"
	endpointVillageProfile = "/game.php?screen=info_village&id=%d"
)

type messageData struct {
	world            string
	date             string
	village          string
	villageURL       string
	oldOwnerName     string
	oldOwnerURL      string
	oldOwnerTribeURL string
	oldOwnerTribeTag string
	newOwnerURL      string
	newOwnerName     string
	newOwnerTribeURL string
	newOwnerTribeTag string
}

type newMessageDataConfig struct {
	host        string
	world       string
	ennoblement *shared_models.Ennoblement
	timezone    string
}

func newMessageData(cfg newMessageDataConfig) messageData {
	data := messageData{
		date:             formatDateOfConquest(getLocation(cfg.timezone), cfg.ennoblement.EnnobledAt),
		world:            cfg.world,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
	}
	if !isVillageNil(cfg.ennoblement.Village) {
		data.village = fmt.Sprintf("%s (%d|%d)",
			cfg.ennoblement.Village.Name,
			cfg.ennoblement.Village.X,
			cfg.ennoblement.Village.Y)
		data.villageURL = utils.FormatVillageURL(cfg.world, cfg.host, cfg.ennoblement.VillageID)
	}
	if !isPlayerNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerName = cfg.ennoblement.OldOwner.Name
		data.oldOwnerURL = utils.FormatPlayerURL(cfg.world, cfg.host, cfg.ennoblement.OldOwner.ID)
	}
	if !isPlayerTribeNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerTribeTag = cfg.ennoblement.OldOwner.Tribe.Tag
		data.oldOwnerTribeURL = utils.FormatTribeURL(cfg.world, cfg.host, cfg.ennoblement.OldOwner.Tribe.ID)
	}
	if !isPlayerNil(cfg.ennoblement.NewOwner) {
		data.newOwnerName = cfg.ennoblement.NewOwner.Name
		data.newOwnerURL = utils.FormatPlayerURL(cfg.world, cfg.host, cfg.ennoblement.NewOwner.ID)
	}
	if !isPlayerTribeNil(cfg.ennoblement.NewOwner) {
		data.newOwnerTribeTag = cfg.ennoblement.NewOwner.Tribe.Tag
		data.newOwnerTribeURL = utils.FormatTribeURL(cfg.world, cfg.host, cfg.ennoblement.NewOwner.Tribe.ID)
	}
	return data
}

func formatMsgLink(text string, url string) string {
	if url == "" {
		return text
	}
	return fmt.Sprintf("[%s](%s)", text, url)
}

func formatMsgAboutVillageLost(msgData messageData) string {
	return fmt.Sprintf(`Wioska %s gracza %s (%s) została stracona na rzecz %s (%s)`,
		formatMsgLink(msgData.village, msgData.villageURL),
		formatMsgLink(msgData.oldOwnerName, msgData.oldOwnerURL),
		formatMsgLink(msgData.oldOwnerTribeTag, msgData.oldOwnerTribeURL),
		formatMsgLink(msgData.newOwnerName, msgData.newOwnerURL),
		formatMsgLink(msgData.newOwnerTribeTag, msgData.newOwnerTribeURL))
}

func formatMsgAboutVillageConquest(msgData messageData) string {
	return fmt.Sprintf(`Gracz %s (%s) podbił wioskę %s od gracza %s (%s)`,
		formatMsgLink(msgData.newOwnerName, msgData.newOwnerURL),
		formatMsgLink(msgData.newOwnerTribeTag, msgData.newOwnerTribeURL),
		formatMsgLink(msgData.village, msgData.villageURL),
		formatMsgLink(msgData.oldOwnerName, msgData.oldOwnerURL),
		formatMsgLink(msgData.oldOwnerTribeTag, msgData.oldOwnerTribeURL))
}
