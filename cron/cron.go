package cron

import (
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

func Attach(c *cron.Cron, cfg Config) {
	h := &handler{
		since:      time.Now().Add(-45 * time.Minute),
		serverRepo: cfg.ServerRepo,
		tribeRepo:  cfg.TribeRepo,
		discord:    cfg.Discord,
		api:        cfg.API,
	}
	c.AddFunc("@every 1m", h.checkLastEnnoblements)
}
