package cron

import (
	"context"
	"log"
	"time"

	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/tribalwarshelp/dcbot/discord"
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
	discord           *discord.Session
	api               *sdk.SDK
}

func (h *handler) loadEnnoblements(worlds []string) map[string]ennoblements {
	m := make(map[string]ennoblements)

	for _, w := range worlds {
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

func (h *handler) loadLangVersions(worlds []string) map[shared_models.LanguageTag]*shared_models.LangVersion {
	languageTags := []shared_models.LanguageTag{}
	cache := make(map[shared_models.LanguageTag]bool)
	for _, world := range worlds {
		languageTag := utils.LanguageTagFromWorldName(world)
		if languageTag.IsValid() && !cache[languageTag] {
			cache[languageTag] = true
			languageTags = append(languageTags, languageTag)
		}
	}

	langVersions := make(map[shared_models.LanguageTag]*shared_models.LangVersion)
	langVersionsList, err := h.api.LangVersions.Browse(&shared_models.LangVersionFilter{
		Tag: languageTags,
	})
	if err == nil {
		for _, langVersion := range langVersionsList.Items {
			langVersions[langVersion.Tag] = langVersion
		}
	} else {
		log.Printf("Cannot load lang versions: %s", err.Error())
	}

	return langVersions
}

func (h *handler) checkLastEnnoblements() {
	worlds, err := h.observationRepo.FetchWorlds(context.Background())
	if err != nil {
		log.Print("checkLastEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: worlds: ", worlds)

	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkLastEnnoblements error: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: number of loaded discord servers: ", total)

	langVersions := h.loadLangVersions(worlds)
	ennoblementsByServerKey := h.loadEnnoblements(worlds)

	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Observations {
			ennoblements, ok := ennoblementsByServerKey[tribe.World]
			langVersion, ok2 := langVersions[utils.LanguageTagFromWorldName(tribe.World)]
			if ok && ok2 {
				if server.LostVillagesChannelID != "" {
					for _, ennoblement := range ennoblements.getLostVillagesByTribe(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.NewOwner) &&
							server.Observations.Contains(tribe.World, ennoblement.NewOwner.Tribe.ID) {
							continue
						}
						newMsgDataConfig := newMessageDataConfig{
							host:        langVersion.Host,
							world:       tribe.World,
							ennoblement: ennoblement,
						}
						msgData := newMessageData(newMsgDataConfig)
						h.discord.SendEmbed(server.LostVillagesChannelID,
							discord.
								NewEmbed().
								SetTitle("Stracona wioska").
								AddField(msgData.world, formatMsgAboutVillageLost(msgData)).
								SetTimestamp(msgData.date).
								MessageEmbed)
					}
				}

				if server.ConqueredVillagesChannelID != "" {
					for _, ennoblement := range ennoblements.getConqueredVillagesByTribe(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.OldOwner) &&
							server.Observations.Contains(tribe.World, ennoblement.OldOwner.Tribe.ID) {
							continue
						}
						newMsgDataConfig := newMessageDataConfig{
							host:        langVersion.Host,
							world:       tribe.World,
							ennoblement: ennoblement,
						}
						msgData := newMessageData(newMsgDataConfig)
						h.discord.SendEmbed(server.ConqueredVillagesChannelID,
							discord.
								NewEmbed().
								SetTitle("Podbita wioska").
								AddField(msgData.world, formatMsgAboutVillageConquest(msgData)).
								SetTimestamp(msgData.date).
								MessageEmbed)
					}
				}
			}
		}
	}
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
	servers, err := h.observationRepo.FetchWorlds(context.Background())
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
			World: keys,
		})
		if err != nil {
			log.Print("deleteClosedTribalWarsServers error: " + err.Error())
		} else {
			log.Printf("deleteClosedTribalWarsServers: total number of deleted observations: %d", len(deleted))
		}
	}
}
