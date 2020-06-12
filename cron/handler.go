package cron

import (
	"context"
	"log"
	"time"

	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/tribe"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

type handler struct {
	lastEnnobledAt map[string]time.Time
	serverRepo     server.Repository
	tribeRepo      tribe.Repository
	discord        *discord.Session
	api            *sdk.SDK
}

func (h *handler) loadEnnoblements(worlds []string) map[string]ennoblements {
	m := make(map[string]ennoblements)

	for _, w := range worlds {
		es, err := h.api.Ennoblements.Browse(w, &sdk.EnnoblementInclude{
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

		lastEnnobledAt, ok := h.lastEnnobledAt[w]
		if !ok {
			lastEnnobledAt = time.Now()
		}

		m[w] = filterEnnoblements(es, lastEnnobledAt)

		lastEnnoblement := m[w].getLastEnnoblement()
		if lastEnnoblement != nil {
			lastEnnobledAt = lastEnnoblement.EnnobledAt.In(time.UTC)
		}
		h.lastEnnobledAt[w] = lastEnnobledAt
	}

	return m
}

func (h *handler) loadLangVersions(worlds []string) map[shared_models.LanguageTag]*shared_models.LangVersion {
	languageTags := []shared_models.LanguageTag{}
	cache := make(map[shared_models.LanguageTag]bool)
	for _, world := range worlds {
		languageTag := utils.LanguageCodeFromWorldName(world)
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
	worlds, err := h.tribeRepo.FetchWorlds(context.Background())
	if err != nil {
		log.Print("checkLastEnnoblements: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: worlds: ", worlds)

	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkLastEnnoblements: " + err.Error())
		return
	}
	log.Print("checkLastEnnoblements: total number of discord servers: ", total)

	langVersions := h.loadLangVersions(worlds)

	ennoblements := h.loadEnnoblements(worlds)
	log.Println("checkLastEnnoblements: loaded ennoblements from", len(ennoblements), "tribalwars servers")

	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Tribes {
			es, ok := ennoblements[tribe.World]
			langVersion, ok2 := langVersions[utils.LanguageCodeFromWorldName(tribe.World)]
			if ok && ok2 {
				if server.LostVillagesChannelID != "" {
					for _, ennoblement := range es.tribeLostVillages(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.NewOwner) &&
							server.Tribes.Contains(tribe.World, ennoblement.NewOwner.Tribe.ID) {
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
					for _, ennoblement := range es.tribeConqueredVillages(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.OldOwner) &&
							server.Tribes.Contains(tribe.World, ennoblement.OldOwner.Tribe.ID) {
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
