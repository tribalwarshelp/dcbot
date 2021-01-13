package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/models"
)

type Command string

func (cmd Command) String() string {
	return string(cmd)
}

func (cmd Command) WithPrefix(prefix string) Command {
	return Command(prefix + cmd.String())
}

type commandCtx struct {
	server    *models.Server
	localizer *i18n.Localizer
}

type commandHandler struct {
	cmd                   Command
	requireAdmPermissions bool
	fn                    func(ctx *commandCtx, m *discordgo.MessageCreate, args ...string)
}

type commandHandlers []*commandHandler

func (hs commandHandlers) find(cmd Command) *commandHandler {
	for _, h := range hs {
		if h.cmd == cmd {
			return h
		}
	}
	return nil
}
