package cron

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

type messageType string

const (
	messageTypeConquer    messageType = "conquer"
	messageTypeLost       messageType = "lost"
	colorLostVillage                  = 0xff0000
	colorConqueredVillage             = 0x00ff00
)

type message struct {
	t                messageType
	server           string
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

type newMessageConfig struct {
	t           messageType
	host        string
	server      string
	ennoblement *shared_models.LiveEnnoblement
}

func newMessage(cfg newMessageConfig) message {
	data := message{
		t:                cfg.t,
		date:             formatDateOfConquest(cfg.ennoblement.EnnobledAt),
		server:           cfg.server,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
	}
	if !isVillageNil(cfg.ennoblement.Village) {
		data.village = fmt.Sprintf("%s (%d|%d) %s",
			cfg.ennoblement.Village.Name,
			cfg.ennoblement.Village.X,
			cfg.ennoblement.Village.Y,
			cfg.ennoblement.Village.Continent())
		data.villageURL = utils.FormatVillageURL(cfg.server, cfg.host, cfg.ennoblement.Village.ID)
	}
	if !isPlayerNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerName = cfg.ennoblement.OldOwner.Name
		data.oldOwnerURL = utils.FormatPlayerURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.ID)
	}
	if !isPlayerTribeNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerTribeTag = cfg.ennoblement.OldOwner.Tribe.Tag
		data.oldOwnerTribeURL = utils.FormatTribeURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.Tribe.ID)
	}
	if !isPlayerNil(cfg.ennoblement.NewOwner) {
		data.newOwnerName = cfg.ennoblement.NewOwner.Name
		data.newOwnerURL = utils.FormatPlayerURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.ID)
	}
	if !isPlayerTribeNil(cfg.ennoblement.NewOwner) {
		data.newOwnerTribeTag = cfg.ennoblement.NewOwner.Tribe.Tag
		data.newOwnerTribeURL = utils.FormatTribeURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.Tribe.ID)
	}

	return data
}

func (msg message) formatMsgAboutVillageLost() string {
	return fmt.Sprintf(`Wioska %s gracza %s (%s) została stracona na rzecz %s (%s)`,
		formatMsgLink(msg.village, msg.villageURL),
		formatMsgLink(msg.oldOwnerName, msg.oldOwnerURL),
		formatMsgLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL),
		formatMsgLink(msg.newOwnerName, msg.newOwnerURL),
		formatMsgLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL))
}

func (msg message) formatMsgAboutVillageConquest() string {
	return fmt.Sprintf("Gracz %s (%s) podbił wioskę %s od gracza %s (%s)",
		formatMsgLink(msg.newOwnerName, msg.newOwnerURL),
		formatMsgLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL),
		formatMsgLink(msg.village, msg.villageURL),
		formatMsgLink(msg.oldOwnerName, msg.oldOwnerURL),
		formatMsgLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL))
}

func (msg message) toEmbed() *discordgo.MessageEmbed {
	title := "Podbita wioska"
	fieldContent := msg.formatMsgAboutVillageConquest()
	color := colorConqueredVillage
	if msg.t == messageTypeLost {
		title = "Stracona wioska"
		fieldContent = msg.formatMsgAboutVillageLost()
		color = colorLostVillage
	}

	return discord.
		NewEmbed().
		SetTitle(title).
		AddField(msg.server, fieldContent).
		SetTimestamp(msg.date).
		SetColor(color).
		MessageEmbed
}
