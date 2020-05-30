package cron

import (
	"context"
	"fmt"
	"log"
	"time"
	"twdcbot/discord"
	"twdcbot/scraper"
	"twdcbot/server"
	"twdcbot/tribe"

	"github.com/robfig/cron/v3"
)

type Config struct {
	ServerRepo server.Repository
	TribeRepo  tribe.Repository
	Discord    *discord.Session
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
	c.AddFunc("@every 1m", h.checkConquers)
}

func (h *handler) checkConquers() {
	worlds, err := h.tribeRepo.FetchWorlds(context.Background())
	if err != nil {
		log.Print("checkConquers: " + err.Error())
		return
	}
	log.Print("checkConquers: worlds: ", worlds)
	servers, total, err := h.serverRepo.Fetch(context.Background(), nil)
	if err != nil {
		log.Print("checkConquers: " + err.Error())
		return
	}
	log.Print("checkConquers: total number of servers: ", total)
	data := scraper.New(worlds, h.since).Scrap()
	h.since = time.Now()
	log.Print("checkConquers: scrapped data: ", data)
	for _, server := range servers {
		if server.ConqueredVillagesChannelID == "" && server.LostVillagesChannelID == "" {
			continue
		}
		for _, tribe := range server.Tribes {
			conquers, ok := data[tribe.World]
			if ok {
				if server.LostVillagesChannelID != "" {
					for _, conquer := range conquers.LostVillages(tribe.TribeID) {
						if server.Tribes.Contains(tribe.World, conquer.NewOwnerTribeID) {
							continue
						}
						h.discord.SendMessage(server.LostVillagesChannelID, formatMsgAboutVillageLost(conquer))
					}
				}
				if server.ConqueredVillagesChannelID != "" {
					for _, conquer := range conquers.ConqueredVillages(tribe.TribeID) {
						if server.Tribes.Contains(tribe.World, conquer.OldOwnerTribeID) {
							continue
						}
						h.discord.SendMessage(server.ConqueredVillagesChannelID, formatMsgAboutVillageConquered(conquer))
					}
				}
			}
		}
	}
}

func formatMsgAboutVillageLost(conquer *scraper.Conquer) string {
	return fmt.Sprintf(`Wioska %s (właściciel: %s [%s]) została stracona na rzecz gracza %s (%s)`,
		conquer.Village,
		conquer.OldOwnerName,
		conquer.OldOwnerTribeName,
		conquer.NewOwnerName,
		conquer.NewOwnerTribeName)
}

func formatMsgAboutVillageConquered(conquer *scraper.Conquer) string {
	return fmt.Sprintf(`Gracz %s (%s) podbił wioskę %s od gracza %s (%s)`,
		conquer.NewOwnerName,
		conquer.NewOwnerTribeName,
		conquer.Village,
		conquer.OldOwnerName,
		conquer.OldOwnerTribeName)
}
