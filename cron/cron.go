package cron

import (
	"time"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/robfig/cron/v3"
)

type Config struct {
	ServerRepo      server.Repository
	ObservationRepo observation.Repository
	Discord         *discord.Session
	GroupRepo       group.Repository
	API             *sdk.SDK
}

func Attach(c *cron.Cron, cfg Config) {
	h := &handler{
		lastEnnoblementAt: make(map[string]time.Time),
		serverRepo:        cfg.ServerRepo,
		observationRepo:   cfg.ObservationRepo,
		groupRepo:         cfg.GroupRepo,
		discord:           cfg.Discord,
		api:               cfg.API,
	}
	c.AddFunc("@every 1m", h.checkLastEnnoblements)
	c.AddFunc("@every 30m", h.checkBotMembershipOnServers)
	c.AddFunc("@every 2h", h.deleteClosedTribalWarsServers)
	go func() {
		h.checkBotMembershipOnServers()
		h.deleteClosedTribalWarsServers()
	}()
}
