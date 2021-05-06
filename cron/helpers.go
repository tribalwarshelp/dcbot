package cron

import (
	"github.com/sirupsen/logrus"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"time"

	"github.com/tribalwarshelp/dcbot/utils"
)

func isBarbarian(p *twmodel.Player) bool {
	return utils.IsPlayerNil(p) || p.ID == 0
}

func trackDuration(log *logrus.Entry, fn func(), fnName string) func() {
	return func() {
		now := time.Now()
		log := log.WithField("fnName", fnName)
		log.Infof("'%s' has been called", fnName)

		fn()

		duration := time.Since(now)
		log.
			WithField("duration", duration.Nanoseconds()).
			WithField("durationPretty", duration.String()).
			Infof("'%s' finished executing", fnName)
	}
}
