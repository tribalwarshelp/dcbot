package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/scraper"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/tribalwars"
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
}

func AttachHandlers(c *cron.Cron, cfg Config) {
	h := &handler{
		since:      time.Now(),
		serverRepo: cfg.ServerRepo,
		tribeRepo:  cfg.TribeRepo,
		discord:    cfg.Discord,
	}
	c.AddFunc("@every 1m", h.checkEnnoblements)
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
	data := scraper.New(worlds, h.since).Scrap()
	h.since = time.Now()
	log.Print("checkEnnoblements: scrapped data: ", data)
	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Tribes {
			conquests, ok := data[tribe.World]
			if ok {
				if server.LostVillagesChannelID != "" {
					for _, conquest := range conquests.LostVillages(tribe.TribeID) {
						if server.Tribes.Contains(tribe.World, conquest.NewOwnerTribeID) {
							continue
						}
						h.discord.SendMessage(server.LostVillagesChannelID, formatMsgAboutVillageLost(tribe.World, conquest))
					}
				}
				if server.ConqueredVillagesChannelID != "" {
					for _, conquest := range conquests.ConqueredVillages(tribe.TribeID) {
						if server.Tribes.Contains(tribe.World, conquest.OldOwnerTribeID) {
							continue
						}
						h.discord.SendMessage(server.ConqueredVillagesChannelID, formatMsgAboutVillageConquest(tribe.World, conquest))
					}
				}
			}
		}
	}
}

func formatDateOfConquest(loc *time.Location, t time.Time) string {
	return t.In(loc).Format("15:04:05")
}

func formatMsgAboutVillageLost(world string, conquest *scraper.Conquest) string {
	return fmt.Sprintf(`**%s** %s: Wioska %s (właściciel: %s [%s]) została stracona na rzecz gracza %s (%s)`,
		world,
		formatDateOfConquest(utils.GetLocation(tribalwars.LanguageCodeFromWorldName(world)), conquest.ConqueredAt),
		conquest.Village,
		conquest.OldOwnerName,
		conquest.OldOwnerTribeName,
		conquest.NewOwnerName,
		conquest.NewOwnerTribeName)
}

func formatMsgAboutVillageConquest(world string, conquest *scraper.Conquest) string {
	return fmt.Sprintf(`**%s** %s: Gracz %s (%s) podbił wioskę %s od gracza %s (%s)`,
		world,
		formatDateOfConquest(utils.GetLocation(tribalwars.LanguageCodeFromWorldName(world)), conquest.ConqueredAt),
		conquest.NewOwnerName,
		conquest.NewOwnerTribeName,
		conquest.Village,
		conquest.OldOwnerName,
		conquest.OldOwnerTribeName)
}
