package discord

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/models"
)

type Command string

func (cmd Command) String() string {
	return string(cmd)
}

func (cmd Command) WithPrefix(prefix string) string {
	return prefix + cmd.String()
}

type commandCtx struct {
	server    *models.Server
	localizer *i18n.Localizer
}
