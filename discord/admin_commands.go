package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/tribalwarshelp/dcbot/message"
)

const (
	cmdChangeLanguage command = "changelanguage"
)

type hndlrChangeLanguage struct {
	*Session
}

var _ commandHandler = &hndlrChangeLanguage{}

func (hndlr *hndlrChangeLanguage) cmd() command {
	return cmdChangeLanguage
}

func (hndlr *hndlrChangeLanguage) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrChangeLanguage) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpChangageLanguage,
				TemplateData: map[string]interface{}{
					"Command":   hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"Languages": getAvailableLanguages(),
				},
			}))
		return
	}

	lang := args[0]
	valid := isValidLanguageTag(lang)
	if !valid {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ChangeLanguageLanguageNotSupported,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	ctx.server.Lang = lang
	if err := hndlr.cfg.ServerRepository.Update(context.Background(), ctx.server); err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	ctx.localizer = message.NewLocalizer(lang)

	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ChangeLanguageSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}
