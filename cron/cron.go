package cron

import (
	"time"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/robfig/cron/v3"
)

type Config struct {
	ServerRepo      server.Repository
	ObservationRepo observation.Repository
	Discord         *discord.Session
	API             *sdk.SDK
}

func Attach(c *cron.Cron, cfg Config) {
	h := &handler{
		lastEnnobledAt:  make(map[string]time.Time),
		serverRepo:      cfg.ServerRepo,
		observationRepo: cfg.ObservationRepo,
		discord:         cfg.Discord,
		api:             cfg.API,
	}
	c.AddFunc("@every 1m", h.checkLastEnnoblements)
	go h.checkBotMembershipOnServers()
	c.AddFunc("@every 30m", h.checkBotMembershipOnServers)
	go h.deleteClosedTribalwarsWorlds()
	c.AddFunc("@every 2h", h.deleteClosedTribalwarsWorlds)
}
