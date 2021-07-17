package cron

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"github.com/tribalwarshelp/shared/tw/twurlbuilder"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/util/twutil"
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
	ennoblement *twmodel.Ennoblement
	localizer   *i18n.Localizer
}

func newMessage(cfg newMessageConfig) checkEnnoblementsMsg {
	data := checkEnnoblementsMsg{
		t:                cfg.t,
		server:           cfg.server,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
		localizer:        cfg.localizer,
	}
	if !twutil.IsVillageNil(cfg.ennoblement.Village) {
		data.village = cfg.ennoblement.Village.FullName()
		data.villageURL = twurlbuilder.BuildVillageURL(cfg.server, cfg.host, cfg.ennoblement.Village.ID)
	}
	if !twutil.IsPlayerNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerName = cfg.ennoblement.OldOwner.Name
		data.oldOwnerURL = twurlbuilder.BuildPlayerURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.ID)
	}
	if !twutil.IsPlayerTribeNil(cfg.ennoblement.OldOwner) {
		data.oldOwnerTribeTag = cfg.ennoblement.OldOwner.Tribe.Tag
		data.oldOwnerTribeURL = twurlbuilder.BuildTribeURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.Tribe.ID)
	}
	if !twutil.IsPlayerNil(cfg.ennoblement.NewOwner) {
		data.newOwnerName = cfg.ennoblement.NewOwner.Name
		data.newOwnerURL = twurlbuilder.BuildPlayerURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.ID)
	}
	if !twutil.IsPlayerTribeNil(cfg.ennoblement.NewOwner) {
		data.newOwnerTribeTag = cfg.ennoblement.NewOwner.Tribe.Tag
		data.newOwnerTribeURL = twurlbuilder.BuildTribeURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.Tribe.ID)
	}

	return data
}

func (msg checkEnnoblementsMsg) String() string {
	return msg.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.CronCheckEnnoblementsMsgLine,
		DefaultMessage: message.FallbackMsg(message.CronCheckEnnoblementsMsgLine,
			"{{.NewOwner}} ({{.NewOwnerTribe}}) has conquered {{.Village}} (Old owner: {{.OldOwner}} ({{.OldOwnerTribe}}))"),
		TemplateData: map[string]interface{}{
			"NewOwner":      discord.BuildLink(msg.newOwnerName, msg.newOwnerURL),
			"NewOwnerTribe": discord.BuildLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL),
			"Village":       discord.BuildLink(msg.village, msg.villageURL),
			"OldOwner":      discord.BuildLink(msg.oldOwnerName, msg.oldOwnerURL),
			"OldOwnerTribe": discord.BuildLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL),
		},
	}) + "\n"
}
