package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/tribalwarshelp/dcbot/model"
)

type command string

func (cmd command) String() string {
	return string(cmd)
}

func (cmd command) WithPrefix(prefix string) string {
	return prefix + cmd.String()
}

type commandCtx struct {
	server    *model.Server
	localizer *i18n.Localizer
}

type commandHandler interface {
	cmd() command
	requireAdmPermissions() bool
	execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string)
}

type commandHandlers []commandHandler

func (hs commandHandlers) find(prefix, cmd string) commandHandler {
	for _, h := range hs {
		if h.cmd().WithPrefix(prefix) == cmd {
			return h
		}
	}
	return nil
}
