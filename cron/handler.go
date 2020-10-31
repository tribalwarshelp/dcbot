package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/tribalwarshelp/shared/tw"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/message"

	"github.com/pkg/errors"
	"github.com/tribalwarshelp/shared/mode"
	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

type handler struct {
	lastEnnoblementAt map[string]time.Time
	serverRepo        server.Repository
	observationRepo   observation.Repository
	groupRepo         group.Repository
	discord           *discord.Session
	api               *sdk.SDK
	status            string
}

func (h *handler) loadEnnoblements(servers []string) (map[string]ennoblements, error) {
	m := make(map[string]ennoblements)

	if len(servers) == 0 {
		return m, nil
	}

	query := ""

	for _, w := range servers {
		query += fmt.Sprintf(`
			%s: liveEnnoblements(server: "%s") {
				%s
				ennobledAt
			}
		`, w, w, sdk.LiveEnnoblementInclude{
			NewOwner: true,
			Village:  true,
			NewOwnerInclude: sdk.PlayerInclude{
				Tribe: true,
			},
			OldOwner: true,
			OldOwnerInclude: sdk.PlayerInclude{
				Tribe: true,
			},
		}.String())
	}

	resp := make(map[string]ennoblements)

	if err := h.api.Post(fmt.Sprintf(`query { %s }`, query), &resp); err != nil {
		return m, errors.Wrap(err, "loadEnnoblements")
	}

	for server, ennoblements := range resp {
		lastEnnoblementAt, ok := h.lastEnnoblementAt[server]
		if !ok {
			lastEnnoblementAt = time.Now().Add(-1 * time.Minute)
		}
		if mode.Get() == mode.DevelopmentMode {
			lastEnnoblementAt = time.Now().Add(-60 * time.Minute)
		}

		m[server] = filterEnnoblements(ennoblements, lastEnnoblementAt)

		lastEnnoblement := m[server].getLastEnnoblement()
		if lastEnnoblement != nil {
			lastEnnoblementAt = lastEnnoblement.EnnobledAt
		}
		h.lastEnnoblementAt[server] = lastEnnoblementAt
	}

	return m, nil
}

func (h *handler) loadLangVersions(servers []string) ([]*shared_models.LangVersion, error) {
	languageTags := []shared_models.LanguageTag{}
	cache := make(map[shared_models.LanguageTag]bool)
	for _, server := range servers {
		languageTag := tw.LanguageTagFromServerKey(server)
		if languageTag.IsValid() && !cache[languageTag] {
			cache[languageTag] = true
			languageTags = append(languageTags, languageTag)
		}
	}

	if len(languageTags) == 0 {
		return []*shared_models.LangVersion{}, nil
	}

	langVersionList, err := h.api.LangVersions.Browse(&shared_models.LangVersionFilter{
		Tag: languageTags,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load lang versions")
	}

	return langVersionList.Items, nil
}

func (h *handler) checkEnnoblements() {
	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("checkEnnoblements: loaded servers")

	groups, total, err := h.groupRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Errorln("checkEnnoblements:", err.Error())
		return
	}
	log.
		WithField("numberOfGroups", total).
		Info("checkEnnoblements: Loaded groups")

	langVersions, err := h.loadLangVersions(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
		return
	}
	log.
		WithField("numberOfLangVersions", len(langVersions)).
		Info("checkEnnoblements: Loaded lang versions")

	ennoblementsByServerKey, err := h.loadEnnoblements(servers)
	if err != nil {
		log.Errorln("checkEnnoblements:", err)
	}
	log.Info("checkEnnoblements: Loaded ennoblements")

	for _, group := range groups {
		if group.ConqueredVillagesChannelID == "" && group.LostVillagesChannelID == "" {
			continue
		}
		localizer := message.NewLocalizer(group.Server.Lang)
		lostVillagesMsg := &discord.MessageEmbed{}
		conqueredVillagesMsg := &discord.MessageEmbed{}
		for _, observation := range group.Observations {
			ennoblements, ok := ennoblementsByServerKey[observation.Server]
			langVersion := utils.FindLangVersionByTag(langVersions, tw.LanguageTagFromServerKey(observation.Server))
			if ok && langVersion != nil && langVersion.Host != "" {
				if group.LostVillagesChannelID != "" {
					for _, ennoblement := range ennoblements.getLostVillagesByTribe(observation.TribeID) {
						if !utils.IsPlayerTribeNil(ennoblement.NewOwner) &&
							group.Observations.Contains(observation.Server, ennoblement.NewOwner.Tribe.ID) {
							continue
						}
						newMsgDataConfig := newMessageConfig{
							host:        langVersion.Host,
							server:      observation.Server,
							ennoblement: ennoblement,
							t:           messageTypeLost,
							localizer:   localizer,
						}
						lostVillagesMsg.Append(newMessage(newMsgDataConfig).String())
					}
				}

				if group.ConqueredVillagesChannelID != "" {
					for _, ennoblement := range ennoblements.getConqueredVillagesByTribe(observation.TribeID, group.ShowInternals) {
						isInTheSameGroup := !utils.IsPlayerTribeNil(ennoblement.OldOwner) &&
							group.Observations.Contains(observation.Server, ennoblement.OldOwner.Tribe.ID)
						if (!group.ShowInternals && isInTheSameGroup) ||
							(!group.ShowEnnobledBarbarians && isBarbarian(ennoblement.OldOwner)) {
							continue
						}

						newMsgDataConfig := newMessageConfig{
							host:        langVersion.Host,
							server:      observation.Server,
							ennoblement: ennoblement,
							t:           messageTypeConquer,
							localizer:   localizer,
						}
						conqueredVillagesMsg.Append(newMessage(newMsgDataConfig).String())
					}
				}
			}
		}

		if group.ConqueredVillagesChannelID != "" && !conqueredVillagesMsg.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronConqueredVillagesTitle,
				DefaultMessage: message.FallbackMsg(message.CronConqueredVillagesTitle,
					"Conquered villages"),
			})
			go h.discord.SendEmbed(group.ConqueredVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorConqueredVillages).
					SetFields(conqueredVillagesMsg.ToMessageEmbedFields()).
					SetTimestamp(formatDateOfConquest(time.Now())).
					MessageEmbed)
		}

		if group.LostVillagesChannelID != "" && !lostVillagesMsg.IsEmpty() {
			title := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CronLostVillagesTitle,
				DefaultMessage: message.FallbackMsg(message.CronLostVillagesTitle,
					"Lost villages"),
			})
			go h.discord.SendEmbed(group.LostVillagesChannelID,
				discord.
					NewEmbed().
					SetTitle(title).
					SetColor(colorLostVillages).
					SetFields(lostVillagesMsg.ToMessageEmbedFields()).
					SetTimestamp(formatDateOfConquest(time.Now())).
					MessageEmbed)
		}
	}
}

func (h *handler) checkBotServers() {
	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Error("checkBotServers: " + err.Error())
		return
	}
	log.
		WithField("numberOfServers", total).
		Info("checkBotServers: loaded servers")

	idsToDelete := []string{}
	for _, server := range servers {
		if isGuildMember, _ := h.discord.IsGuildMember(server.ID); !isGuildMember {
			idsToDelete = append(idsToDelete, server.ID)
		}
	}

	if len(idsToDelete) > 0 {
		deleted, err := h.serverRepo.Delete(context.Background(), &models.ServerFilter{
			ID: idsToDelete,
		})
		if err != nil {
			log.Error("checkBotServers: " + err.Error())
		} else {
			log.
				WithField("numberOfDeletedServers", len(deleted)).
				Info("checkBotServers: deleted servers")
		}
	}
}

func (h *handler) deleteClosedTribalWarsServers() {
	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Error("deleteClosedTribalWarsServers: " + err.Error())
		return
	}
	log.
		WithField("servers", servers).
		Info("deleteClosedTribalWarsServers: loaded servers")

	list, err := h.api.Servers.Browse(&shared_models.ServerFilter{
		Key:    servers,
		Status: []shared_models.ServerStatus{shared_models.ServerStatusClosed},
	}, nil)
	if err != nil {
		log.Errorln("deleteClosedTribalWarsServers: " + err.Error())
		return
	}
	if list == nil || list.Items == nil {
		return
	}

	keys := []string{}
	for _, server := range list.Items {
		keys = append(keys, server.Key)
	}

	if len(keys) > 0 {
		deleted, err := h.observationRepo.Delete(context.Background(), &models.ObservationFilter{
			Server: keys,
		})
		if err != nil {
			log.Errorln("deleteClosedTribalWarsServers: " + err.Error())
		} else {
			log.
				WithField("numberOfDeletedObservations", len(deleted)).
				Infof("deleteClosedTribalWarsServers: deleted observations")
		}
	}
}

func (h *handler) updateBotStatus() {
	if err := h.discord.UpdateStatus(h.status); err != nil {
		log.Error("updateBotStatus: " + err.Error())
	}
}
