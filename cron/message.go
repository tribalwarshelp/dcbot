package cron

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/message"
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

func (msg checkEnnoblementsMsg) String() string {
	return msg.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "cron.checkEnnoblements.msgLine",
		DefaultMessage: message.FallbackMsg("cron.checkEnnoblements.msgLine",
			"{{.NewOwner}} ({{.NewOwnerTribe}}) has conquered the village {{.Village}} (Old owner: {{.OldOwner}} ({{.OldOwnerTribe}}))"),
		TemplateData: map[string]interface{}{
			"NewOwner":      formatMsgLink(msg.newOwnerName, msg.newOwnerURL),
			"NewOwnerTribe": formatMsgLink(msg.newOwnerTribeTag, msg.newOwnerTribeURL),
			"Village":       formatMsgLink(msg.village, msg.villageURL),
			"OldOwner":      formatMsgLink(msg.oldOwnerName, msg.oldOwnerURL),
			"OldOwnerTribe": formatMsgLink(msg.oldOwnerTribeTag, msg.oldOwnerTribeURL),
		},
	}) + "\n"
}
