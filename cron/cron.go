package cron

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tribalwarshelp/golang-sdk/sdk"
	"github.com/tribalwarshelp/shared/mode"

	"github.com/tribalwarshelp/dcbot/discord"
	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/robfig/cron/v3"
)

var log = logrus.WithField("package", "cron")

type Config struct {
	ServerRepo      server.Repository
	ObservationRepo observation.Repository
	Discord         *discord.Session
	GroupRepo       group.Repository
	API             *sdk.SDK
	Status          string
}

func Attach(c *cron.Cron, cfg Config) {
	w := &worker{
		lastEnnoblementAt: make(map[string]time.Time),
		serverRepo:        cfg.ServerRepo,
		observationRepo:   cfg.ObservationRepo,
		groupRepo:         cfg.GroupRepo,
		discord:           cfg.Discord,
		api:               cfg.API,
		status:            cfg.Status,
	}
	c.AddFunc("@every 1m", w.checkEnnoblements)
	c.AddFunc("@every 30m", w.checkBotServers)
	c.AddFunc("@every 2h10m", w.deleteClosedTribalWarsServers)
	c.AddFunc("@every 2h", w.updateBotStatus)
	go func() {
		w.checkBotServers()
		w.deleteClosedTribalWarsServers()
		w.updateBotStatus()
		if mode.Get() == mode.DevelopmentMode {
			w.checkEnnoblements()
		}
	}()
}
