package cron

import (
	"context"
	"fmt"
	"log"
	"time"

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
		languageTag := utils.LanguageTagFromWorldName(server)
		if languageTag.IsValid() && !cache[languageTag] {
			cache[languageTag] = true
			languageTags = append(languageTags, languageTag)
		}
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
	start := time.Now()

	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Print("checkEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkEnnoblements: servers: ", servers)

	groups, total, err := h.groupRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkEnnoblements: number of loaded groups: ", total)

	langVersions, err := h.loadLangVersions(servers)
	if err != nil {
		log.Print(err)
		return
	}
	ennoblementsByServerKey, err := h.loadEnnoblements(servers)
	if err != nil {
		log.Print(err)
		return
	}

	for _, group := range groups {
		if group.ConqueredVillagesChannelID == "" && group.LostVillagesChannelID == "" {
			continue
		}
		localizer := message.NewLocalizer(group.Server.Lang)
		lostVillagesMsg := &discord.EmbedMessage{}
		conqueredVillagesMsg := &discord.EmbedMessage{}
		for _, observation := range group.Observations {
			ennoblements, ok := ennoblementsByServerKey[observation.Server]
			langVersion := utils.FindLangVersionByTag(langVersions, utils.LanguageTagFromWorldName(observation.Server))
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
				MessageID: "cron.conqueredVillages.title",
				DefaultMessage: message.FallbackMsg("cron.conqueredVillages.title",
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
				MessageID: "cron.lostVillages.title",
				DefaultMessage: message.FallbackMsg("cron.lostVillages.title",
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

	log.Printf("checkEnnoblements: finished in %s", time.Since(start).String())
}

func (h *handler) checkBotServers() {
	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkBotServers error: " + err.Error())
		return
	}
	log.Print("checkBotServers: total number of loaded discord servers: ", total)

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
			log.Print("checkBotServers error: " + err.Error())
		} else {
			log.Printf("checkBotServers: total number of deleted discord servers: %d", len(deleted))
		}
	}
}

func (h *handler) deleteClosedTribalWarsServers() {
	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Print("deleteClosedTribalWarsServers: " + err.Error())
		return
	}
	log.Print("deleteClosedTribalWarsServers: servers: ", servers)

	list, err := h.api.Servers.Browse(&shared_models.ServerFilter{
		Key:    servers,
		Status: []shared_models.ServerStatus{shared_models.ServerStatusClosed},
	}, nil)
	if err != nil {
		log.Print("deleteClosedTribalWarsServers: " + err.Error())
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
			log.Print("deleteClosedTribalWarsServers error: " + err.Error())
		} else {
			log.Printf("deleteClosedTribalWarsServers: total number of deleted observations: %d", len(deleted))
		}
	}
}

func (h *handler) updateBotStatus() {
	if err := h.discord.UpdateStatus(h.status); err != nil {
		log.Print("updateBotStatus: " + err.Error())
	}
}
