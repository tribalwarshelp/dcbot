package discord

import (
	"context"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
	"github.com/tribalwarshelp/shared/models"
	"github.com/tribalwarshelp/shared/tw"
)

const (
	coordsLimit                             = 20
	CoordsTranslationCommand        Command = "coordstranslation"
	DisableCoordsTranslationCommand Command = "disablecoordstranslation"
)

var coordsRegex = regexp.MustCompile(`(\d+)\|(\d+)`)

func (s *Session) handleCoordsTranslationCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpCoordsTranslation,
				DefaultMessage: message.FallbackMsg(message.HelpCoordsTranslation,
					"**{{.Command}}** [server] - enables coords translation feature."),
				TemplateData: map[string]interface{}{
					"Command": CoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	serverKey := args[0]
	server, err := s.cfg.API.Servers.Read(serverKey, nil)
	if err != nil || server == nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.CoordsTranslationServerNotFound,
				DefaultMessage: message.FallbackMsg(message.CoordsTranslationServerNotFound, "{{.Mention}} Server not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	ctx.server.CoordsTranslation = serverKey
	go s.cfg.ServerRepository.Update(context.Background(), ctx.server)

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.CoordsTranslationSuccess,
			DefaultMessage: message.FallbackMsg(message.CoordsTranslationSuccess,
				"{{.Mention}} Coords translation feature has been enabled."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDisableCoordsTranslationCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	ctx.server.CoordsTranslation = ""
	go s.cfg.ServerRepository.Update(context.Background(), ctx.server)

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableCoordsTranslationSuccess,
			DefaultMessage: message.FallbackMsg(message.DisableCoordsTranslationSuccess,
				"{{.Mention}} Coords translation feature has been disabled."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) translateCoords(ctx *commandCtx, m *discordgo.MessageCreate) {
	if ctx.server.CoordsTranslation == "" {
		return
	}
	coords := extractAllCoordsFromMessage(m.Content)
	coordsLen := len(coords)
	if coordsLen > 0 {
		langVersion, err := s.cfg.API.LangVersions.Read(tw.LanguageTagFromServerKey(ctx.server.CoordsTranslation))
		if err != nil || langVersion == nil {
			return
		}
		if coordsLen > coordsLimit {
			coords = coords[0:coordsLimit]
		}
		list, err := s.cfg.API.Villages.Browse(ctx.server.CoordsTranslation,
			&models.VillageFilter{
				XY: coords,
			},
			&sdk.VillageInclude{
				Player: true,
				PlayerInclude: sdk.PlayerInclude{
					Tribe: true,
				},
			})
		if err != nil || list == nil || list.Items == nil || len(list.Items) <= 0 {
			return
		}

		msg := &MessageEmbed{}
		for _, village := range list.Items {
			villageURL := tw.BuildVillageURL(ctx.server.CoordsTranslation, langVersion.Host, village.ID)
			playerName := "-"
			playerURL := ""
			if !utils.IsPlayerNil(village.Player) {
				playerName = village.Player.Name
				playerURL = tw.BuildPlayerURL(ctx.server.CoordsTranslation, langVersion.Host, village.Player.ID)
			}
			tribeName := "-"
			tribeURL := ""
			if !utils.IsPlayerTribeNil(village.Player) {
				tribeName = village.Player.Tribe.Name
				tribeURL = tw.BuildTribeURL(ctx.server.CoordsTranslation, langVersion.Host, village.Player.Tribe.ID)
			}

			msg.Append(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.CoordsTranslationMessage,
				DefaultMessage: message.FallbackMsg(message.CoordsTranslationMessage,
					"{{.Village}} owned by {{.Player}} (Tribe: {{.Tribe}})."),
				TemplateData: map[string]interface{}{
					"Village": BuildLink(village.FullName(), villageURL),
					"Player":  BuildLink(playerName, playerURL),
					"Tribe":   BuildLink(tribeName, tribeURL),
				},
			}) + "\n")
		}

		s.SendEmbed(m.ChannelID, NewEmbed().
			SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.CoordsTranslationTitle,
				DefaultMessage: message.FallbackMsg(message.CoordsTranslationTitle, "Villages"),
			})).
			SetFields(msg.ToMessageEmbedFields()).
			MessageEmbed)
	}
}

func extractAllCoordsFromMessage(msg string) []string {
	coords := []string{}
	for _, bytes := range coordsRegex.FindAll([]byte(msg), -1) {
		coords = append(coords, string(bytes))
	}
	return coords
}
