package discord

import "github.com/bwmarrin/discordgo"

type messageProcessor interface {
	process(ctx *commandCtx, m *discordgo.MessageCreate)
}
