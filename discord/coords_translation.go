package discord

import (
	"context"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"github.com/tribalwarshelp/shared/tw/twurlbuilder"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/util/twutil"
)

const (
	coordsLimit                             = 20
	CoordsTranslationCommand        Command = "coordstranslation"
	DisableCoordsTranslationCommand Command = "disablecoordstranslation"
)

var coordsRegex = regexp.MustCompile(`(\d+)\|(\d+)`)

type commandCoordsTranslation struct {
	*Session
}

var _ commandHandler = &commandCoordsTranslation{}

func (c *commandCoordsTranslation) cmd() Command {
	return CoordsTranslationCommand
}

func (c *commandCoordsTranslation) requireAdmPermissions() bool {
	return true
}

func (c *commandCoordsTranslation) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(
			m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpCoordsTranslation,
				TemplateData: map[string]interface{}{
					"Command": c.cmd().WithPrefix(c.cfg.CommandPrefix),
				},
			}),
		)
		return
	}

	serverKey := args[0]
	server, err := c.cfg.API.Server.Read(serverKey, nil)
	if err != nil || server == nil {
		c.SendMessage(
			m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CoordsTranslationServerNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}),
		)
		return
	}

	ctx.server.CoordsTranslation = serverKey
	go c.cfg.ServerRepository.Update(context.Background(), ctx.server)

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.CoordsTranslationSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandDisableCoordsTranslation struct {
	*Session
}

var _ commandHandler = &commandDisableCoordsTranslation{}

func (c *commandDisableCoordsTranslation) cmd() Command {
	return DisableCoordsTranslationCommand
}

func (c *commandDisableCoordsTranslation) requireAdmPermissions() bool {
	return true
}

func (c *commandDisableCoordsTranslation) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	ctx.server.CoordsTranslation = ""
	go c.cfg.ServerRepository.Update(context.Background(), ctx.server)

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableCoordsTranslationSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) translateCoords(ctx *commandCtx, m *discordgo.MessageCreate) {
	if ctx.server.CoordsTranslation == "" {
		return
	}
	coords := coordsRegex.FindAllString(m.Content, -1)
	coordsLen := len(coords)
	if coordsLen > 0 {
		version, err := s.cfg.API.Version.Read(twmodel.VersionCodeFromServerKey(ctx.server.CoordsTranslation))
		if err != nil || version == nil {
			return
		}
		if coordsLen > coordsLimit {
			coords = coords[0:coordsLimit]
		}
		list, err := s.cfg.API.Village.Browse(ctx.server.CoordsTranslation,
			0,
			0,
			[]string{},
			&twmodel.VillageFilter{
				XY: coords,
			},
			&sdk.VillageInclude{
				Player: true,
				PlayerInclude: sdk.PlayerInclude{
					Tribe: true,
				},
			},
		)
		if err != nil || list == nil || len(list.Items) <= 0 {
			return
		}

		msg := &MessageEmbedFieldBuilder{}
		for _, village := range list.Items {
			villageURL := twurlbuilder.BuildVillageURL(ctx.server.CoordsTranslation, version.Host, village.ID)
			playerName := "-"
			playerURL := ""
			if !twutil.IsPlayerNil(village.Player) {
				playerName = village.Player.Name
				playerURL = twurlbuilder.BuildPlayerURL(ctx.server.CoordsTranslation, version.Host, village.Player.ID)
			}
			tribeName := "-"
			tribeURL := ""
			if !twutil.IsPlayerTribeNil(village.Player) {
				tribeName = village.Player.Tribe.Name
				tribeURL = twurlbuilder.BuildTribeURL(ctx.server.CoordsTranslation, version.Host, village.Player.Tribe.ID)
			}

			msg.Append(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CoordsTranslationMessage,
				TemplateData: map[string]interface{}{
					"Village": BuildLink(village.FullName(), villageURL),
					"Player":  BuildLink(playerName, playerURL),
					"Tribe":   BuildLink(tribeName, tribeURL),
				},
			}) + "\n")
		}

		s.SendEmbed(m.ChannelID, NewEmbed().
			SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CoordsTranslationTitle,
			})).
			SetFields(msg.ToMessageEmbedFields()))
	}
}
