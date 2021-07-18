package cron

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"github.com/tribalwarshelp/shared/tw/twurlbuilder"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/util/twutil"
)

type ennoblementMsg struct {
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

type newEnnoblementMsgConfig struct {
	host        string
	server      string
	ennoblement *twmodel.Ennoblement
	localizer   *i18n.Localizer
}

func newEnnoblementMsg(cfg newEnnoblementMsgConfig) ennoblementMsg {
	msg := ennoblementMsg{
		server:           cfg.server,
		village:          "-",
		oldOwnerName:     "-",
		oldOwnerTribeTag: "-",
		newOwnerName:     "-",
		newOwnerTribeTag: "-",
		localizer:        cfg.localizer,
	}
	if !twutil.IsVillageNil(cfg.ennoblement.Village) {
		msg.village = cfg.ennoblement.Village.FullName()
		msg.villageURL = twurlbuilder.BuildVillageURL(cfg.server, cfg.host, cfg.ennoblement.Village.ID)
	}
	if !twutil.IsPlayerNil(cfg.ennoblement.OldOwner) {
		msg.oldOwnerName = cfg.ennoblement.OldOwner.Name
		msg.oldOwnerURL = twurlbuilder.BuildPlayerURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.ID)
	}
	if !twutil.IsPlayerTribeNil(cfg.ennoblement.OldOwner) {
		msg.oldOwnerTribeTag = cfg.ennoblement.OldOwner.Tribe.Tag
		msg.oldOwnerTribeURL = twurlbuilder.BuildTribeURL(cfg.server, cfg.host, cfg.ennoblement.OldOwner.Tribe.ID)
	}
	if !twutil.IsPlayerNil(cfg.ennoblement.NewOwner) {
		msg.newOwnerName = cfg.ennoblement.NewOwner.Name
		msg.newOwnerURL = twurlbuilder.BuildPlayerURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.ID)
	}
	if !twutil.IsPlayerTribeNil(cfg.ennoblement.NewOwner) {
		msg.newOwnerTribeTag = cfg.ennoblement.NewOwner.Tribe.Tag
		msg.newOwnerTribeURL = twurlbuilder.BuildTribeURL(cfg.server, cfg.host, cfg.ennoblement.NewOwner.Tribe.ID)
	}

	return msg
}

func (msg ennoblementMsg) String() string {
	return msg.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.CronCheckEnnoblementsMsgLine,
		TemplateData: map[string]interface{}{
			"NewOwner":      discord.BuildLink(msg.newOwnerName, msg.newOwnerURL),
			"NewOwnerTribe": discord.BuildLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL),
			"Village":       discord.BuildLink(msg.village, msg.villageURL),
			"OldOwner":      discord.BuildLink(msg.oldOwnerName, msg.oldOwnerURL),
			"OldOwnerTribe": discord.BuildLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL),
		},
	}) + "\n"
}
