package cron

import (
	"github.com/Kichiyaki/appmode"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tribalwarshelp/golang-sdk/sdk"

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
	h := &handler{
		lastEnnoblementAt: make(map[string]time.Time),
		serverRepo:        cfg.ServerRepo,
		observationRepo:   cfg.ObservationRepo,
		groupRepo:         cfg.GroupRepo,
		discord:           cfg.Discord,
		api:               cfg.API,
		status:            cfg.Status,
	}
	checkEnnoblements := trackDuration(log, h.checkEnnoblements, "checkEnnoblements")
	checkBotServers := trackDuration(log, h.checkBotServers, "checkBotServers")
	deleteClosedTribalWarsServers := trackDuration(log,
		h.deleteClosedTWServers,
		"deleteClosedTWServers")
	updateBotStatus := trackDuration(log, h.updateBotStatus, "updateBotStatus")
	c.AddFunc("@every 1m", checkEnnoblements)
	c.AddFunc("@every 30m", checkBotServers)
	c.AddFunc("@every 2h10m", deleteClosedTribalWarsServers)
	c.AddFunc("@every 2h", updateBotStatus)
	go func() {
		checkBotServers()
		deleteClosedTribalWarsServers()
		updateBotStatus()
		if appmode.Equals(appmode.DevelopmentMode) {
			checkEnnoblements()
		}
	}()
}
