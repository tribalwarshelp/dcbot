package cron

import (
	"context"
	"log"
	"time"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/tribe"

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
		since:      time.Now().Add(-30 * time.Minute),
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
	data := h.loadEnnoblements(worlds)
	h.since = time.Now()
	log.Print("checkEnnoblements: scrapped data: ", data)
	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Tribes {
			es, ok := data[tribe.World]
			if ok {
				if server.LostVillagesChannelID != "" {
					for _, ennoblement := range es.tribeLostVillages(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.NewOwner) &&
							server.Tribes.Contains(tribe.World, ennoblement.NewOwner.Tribe.ID) {
							continue
						}
						msgData := newMessageData(tribe.World, ennoblement)
						h.discord.SendMessage(server.LostVillagesChannelID, formatMsgAboutVillageLost(msgData))
					}
				}
				if server.ConqueredVillagesChannelID != "" {
					for _, ennoblement := range es.tribeConqueredVillages(tribe.TribeID) {
						if !isPlayerTribeNil(ennoblement.OldOwner) &&
							server.Tribes.Contains(tribe.World, ennoblement.OldOwner.Tribe.ID) {
							continue
						}
						msgData := newMessageData(tribe.World, ennoblement)
						h.discord.SendMessage(server.ConqueredVillagesChannelID, formatMsgAboutVillageConquest(msgData))
					}
				}
			}
		}
	}
}
