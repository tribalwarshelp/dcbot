package cron

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

type messageType string

const (
	messageTypeConquer     messageType = "conquer"
	messageTypeLost        messageType = "lost"
	colorLostVillages                  = 0xff0000
	colorConqueredVillages             = 0x00ff00
)

type checkEnnoblementsMsg struct {
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
	localizer        *i18n.Localizer
}

type newMessageConfig struct {
	t           messageType
	host        string
	server      string
	ennoblement *shared_models.LiveEnnoblement
	localizer   *i18n.Localizer
}

func newMessage(cfg newMessageConfig) checkEnnoblementsMsg {
	data := checkEnnoblementsMsg{
		t:                cfg.t,
		date:             formatDateOfConquest(cfg.ennoblement.EnnobledAt),
		server:           cfg.server,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
		localizer:        cfg.localizer,
	}
	if !utils.IsVillageNil(cfg.ennoblement.Village) {
		data.village = cfg.ennoblement.Village.FullName()
		data.villageURL = utils.FormatVillageURL(cfg.server, cfg.host, cfg.ennoblement.Village.ID)
	}
	if !utils.IsPlayerNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerName = cfg.ennoblement.OldOwner.Name
		data.oldOwnerURL = utils.FormatPlayerURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.ID)
	}
	if !utils.IsPlayerTribeNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerTribeTag = cfg.ennoblement.OldOwner.Tribe.Tag
		data.oldOwnerTribeURL = utils.FormatTribeURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.Tribe.ID)
	}
	if !utils.IsPlayerNil(cfg.ennoblement.NewOwner) {
		data.newOwnerName = cfg.ennoblement.NewOwner.Name
		data.newOwnerURL = utils.FormatPlayerURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.ID)
	}
	if !utils.IsPlayerTribeNil(cfg.ennoblement.NewOwner) {
		data.newOwnerTribeTag = cfg.ennoblement.NewOwner.Tribe.Tag
		data.newOwnerTribeURL = utils.FormatTribeURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.Tribe.ID)
	}

	return data
}

func (msg checkEnnoblementsMsg) String() string {
	return msg.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "cron.checkEnnoblements.msgLine",
		DefaultMessage: message.FallbackMsg("cron.checkEnnoblements.msgLine",
			"{{.NewOwner}} ({{.NewOwnerTribe}}) conquered {{.Village}} (Old owner: {{.OldOwner}} ({{.OldOwnerTribe}}))"),
		TemplateData: map[string]interface{}{
			"NewOwner":      discord.FormatLink(msg.newOwnerName, msg.newOwnerURL),
			"NewOwnerTribe": discord.FormatLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL),
			"Village":       discord.FormatLink(msg.village, msg.villageURL),
			"OldOwner":      discord.FormatLink(msg.oldOwnerName, msg.oldOwnerURL),
			"OldOwnerTribe": discord.FormatLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL),
		},
	}) + "\n"
}
