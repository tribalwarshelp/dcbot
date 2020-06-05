package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tribalwarshelp/golang-sdk/sdk"
	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/tribalwarshelp/dcbot/discord"
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

func formatDateOfConquest(loc *time.Location, t time.Time) string {
	return t.In(loc).Format("15:04:05")
}

type messageData struct {
	world            string
	date             string
	village          string
	oldOwnerName     string
	oldOwnerTribeTag string
	newOwnerName     string
	newOwnerTribeTag string
}

func newMessageData(world string, ennoblement *shared_models.Ennoblement) messageData {
	data := messageData{
		date:  formatDateOfConquest(utils.GetLocation(tribalwars.LanguageCodeFromWorldName(world)), ennoblement.EnnobledAt),
		world: world,
	}
	if !isVillageNil(ennoblement.Village) {
		data.village = fmt.Sprintf("%s (%d|%d)", ennoblement.Village.Name, ennoblement.Village.X, ennoblement.Village.Y)
	}
	if !isPlayerNil(ennoblement.OldOwner) {
		data.oldOwnerName = ennoblement.OldOwner.Name
	}
	if !isPlayerTribeNil(ennoblement.OldOwner) {
		data.oldOwnerTribeTag = ennoblement.OldOwner.Tribe.Tag
	}
	if !isPlayerNil(ennoblement.NewOwner) {
		data.newOwnerName = ennoblement.NewOwner.Name
	}
	if !isPlayerTribeNil(ennoblement.NewOwner) {
		data.newOwnerTribeTag = ennoblement.NewOwner.Tribe.Tag
	}
	return data
}

func formatMsgAboutVillageLost(msgData messageData) string {

	return fmt.Sprintf(`**%s** %s: Wioska %s (właściciel: %s [%s]) została stracona na rzecz gracza %s (%s)`,
		msgData.world,
		msgData.date,
		msgData.village,
		msgData.oldOwnerName,
		msgData.oldOwnerTribeTag,
		msgData.newOwnerName,
		msgData.newOwnerTribeTag)
}

func formatMsgAboutVillageConquest(msgData messageData) string {
	return fmt.Sprintf(`**%s** %s: Gracz %s (%s) podbił wioskę %s od gracza %s (%s)`,
		msgData.world,
		msgData.date,
		msgData.newOwnerName,
		msgData.newOwnerTribeTag,
		msgData.village,
		msgData.oldOwnerName,
		msgData.oldOwnerTribeTag)
}
