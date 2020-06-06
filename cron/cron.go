package cron

import (
	"context"
	"log"
	"time"

	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/tribe"
	"github.com/tribalwarshelp/dcbot/utils"

	"github.com/robfig/cron/v3"
)

type Config struct {
	ServerRepo server.Repository
	TribeRepo  tribe.Repository
	Discord    *discord.Session
	API        *sdk.SDK
}

type handler struct {
	since      time.Time
	serverRepo server.Repository
	tribeRepo  tribe.Repository
	discord    *discord.Session
	api        *sdk.SDK
}

func AttachHandlers(c *cron.Cron, cfg Config) {
	h := &handler{
		since:      time.Now(),
		serverRepo: cfg.ServerRepo,
		tribeRepo:  cfg.TribeRepo,
		discord:    cfg.Discord,
		api:        cfg.API,
	}
	c.AddFunc("@every 1m", h.checkEnnoblements)
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
		m[w] = filterEnnoblements(es, h.since)
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

func (h *handler) checkEnnoblements() {
	worlds, err := h.tribeRepo.FetchWorlds(context.Background())
	if err != nil {
		log.Print("checkEnnoblements: " + err.Error())
		return
	}
	log.Print("checkEnnoblements: worlds: ", worlds)

	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkEnnoblements: " + err.Error())
		return
	}
	log.Print("checkEnnoblements: total number of servers: ", total)

	langVersions := h.loadLangVersions(worlds)

	data := h.loadEnnoblements(worlds)
	h.since = time.Now()
	log.Print("checkEnnoblements: loaded ennoblements: ", len(data))

	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Tribes {
			es, ok := data[tribe.World]
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
							timezone:    langVersion.Timezone,
						}
						msgData := newMessageData(newMsgDataConfig)
						h.discord.SendEmbed(server.LostVillagesChannelID,
							discord.
								NewEmbed().
								SetTitle("Stracona wioska").
								SetTimestamp(msgData.date).
								AddField(msgData.world, formatMsgAboutVillageLost(msgData)).
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
							timezone:    langVersion.Timezone,
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
