package cron

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
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

func (h *handler) loadEnnoblements(servers []string) map[string]ennoblements {
	m := make(map[string]ennoblements)

	for _, w := range servers {
		es, err := h.api.LiveEnnoblements.Browse(w, &sdk.LiveEnnoblementInclude{
			NewOwner: true,
			Village:  true,
			NewOwnerInclude: sdk.PlayerInclude{
				Tribe: true,
			},
			OldOwner: true,
			OldOwnerInclude: sdk.PlayerInclude{
				Tribe: true,
			},
		})
		if err != nil {
			log.Printf("%s: %s", w, err.Error())
			continue
		}

		lastEnnoblementAt, ok := h.lastEnnoblementAt[w]
		if !ok {
			lastEnnoblementAt = time.Now().Add(-1 * time.Minute)
		}

		m[w] = filterEnnoblements(es, lastEnnoblementAt)

		lastEnnoblement := m[w].getLastEnnoblement()
		if lastEnnoblement != nil {
			lastEnnoblementAt = lastEnnoblement.EnnobledAt
		}
		h.lastEnnoblementAt[w] = lastEnnoblementAt
	}

	return m
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

func (h *handler) checkLastEnnoblements() {
	start := time.Now()

	servers, err := h.observationRepo.FetchServers(context.Background())
	if err != nil {
		log.Print("checkLastEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: servers: ", servers)

	groups, total, err := h.groupRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkLastEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: number of loaded groups: ", total)

	langVersions, err := h.loadLangVersions(servers)
	if err != nil {
		log.Print(err)
		return
	}
	ennoblementsByServerKey := h.loadEnnoblements(servers)

	for _, group := range groups {
		if group.ConqueredVillagesChannelID == "" && group.LostVillagesChannelID == "" {
			continue
		}
		for _, observation := range group.Observations {
			ennoblements, ok := ennoblementsByServerKey[observation.Server]
			langVersion := utils.FindLangVersionByTag(langVersions, utils.LanguageTagFromWorldName(observation.Server))
			if ok && langVersion != nil && langVersion.Host != "" {
				if group.LostVillagesChannelID != "" {
					msg := &discord.EmbedMessage{}
					for _, ennoblement := range ennoblements.getLostVillagesByTribe(observation.TribeID) {
						if !isPlayerTribeNil(ennoblement.NewOwner) &&
							group.Observations.Contains(observation.Server, ennoblement.NewOwner.Tribe.ID) {
							continue
						}
						newMsgDataConfig := newMessageConfig{
							host:        langVersion.Host,
							server:      observation.Server,
							ennoblement: ennoblement,
							t:           messageTypeLost,
						}
						msg.Append(newMessage(newMsgDataConfig).String())
					}
					if !msg.IsEmpty() {
						h.discord.SendEmbed(group.LostVillagesChannelID,
							discord.
								NewEmbed().
								SetTitle("Stracone wioski").
								SetColor(colorLostVillage).
								SetFields(msg.ToMessageEmbedFields()).
								SetTimestamp(formatDateOfConquest(time.Now())).
								MessageEmbed)
					}
				}

				if group.ConqueredVillagesChannelID != "" {
					msg := &discord.EmbedMessage{}
					for _, ennoblement := range ennoblements.getConqueredVillagesByTribe(observation.TribeID) {
						isBarbarian := isPlayerNil(ennoblement.OldOwner) || ennoblement.OldOwner.ID == 0
						if (!isPlayerTribeNil(ennoblement.OldOwner) &&
							group.Observations.Contains(observation.Server, ennoblement.OldOwner.Tribe.ID)) ||
							(!group.ShowEnnobledBarbarians && isBarbarian) {
							continue
						}
						newMsgDataConfig := newMessageConfig{
							host:        langVersion.Host,
							server:      observation.Server,
							ennoblement: ennoblement,
							t:           messageTypeConquer,
						}
						msg.Append(newMessage(newMsgDataConfig).String())
					}
					if !msg.IsEmpty() {
						h.discord.SendEmbed(group.ConqueredVillagesChannelID,
							discord.
								NewEmbed().
								SetTitle("Podbite wioski").
								SetColor(colorConqueredVillage).
								SetFields(msg.ToMessageEmbedFields()).
								SetTimestamp(formatDateOfConquest(time.Now())).
								MessageEmbed)
					}
				}
			}
		}
	}

	log.Printf("checkLastEnnoblements: finished in %s", time.Since(start).String())
}

func (h *handler) checkBotMembershipOnServers() {
	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkBotMembershipOnServers error: " + err.Error())
		return
	}
	log.Print("checkBotMembershipOnServers: total number of loaded discord servers: ", total)

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
			log.Print("checkBotMembershipOnServers error: " + err.Error())
		} else {
			log.Printf("checkBotMembershipOnServers: total number of deleted discord servers: %d", len(deleted))
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
