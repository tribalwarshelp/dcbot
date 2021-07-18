package discord

import (
	"fmt"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"github.com/tribalwarshelp/shared/tw/twurlbuilder"
	"math"
	"strconv"
	"strings"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

const (
	cmdHelp      command = "help"
	cmdTribe     command = "tribe"
	cmdTopODA    command = "topoda"
	cmdTopODD    command = "topodd"
	cmdTopODS    command = "topods"
	cmdTopOD     command = "topod"
	cmdTopPoints command = "toppoints"
	cmdAuthor    command = "author"
)

type hndlrHelp struct {
	*Session
}

var _ commandHandler = &hndlrHelp{}

func (hndlr *hndlrHelp) cmd() command {
	return cmdHelp
}

func (hndlr *hndlrHelp) requireAdmPermissions() bool {
	return false
}

func (hndlr *hndlrHelp) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	commandsForAll := fmt.Sprintf(`
- %s
- %s
- %s
- %s
- %s
- %s
				`,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODA,
			TemplateData: map[string]interface{}{
				"Command": cmdTribe.WithPrefix(hndlr.cfg.CommandPrefix) + " " + cmdTopODA.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODD,
			TemplateData: map[string]interface{}{
				"Command": cmdTribe.WithPrefix(hndlr.cfg.CommandPrefix) + " " + cmdTopODD.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODS,
			TemplateData: map[string]interface{}{
				"Command": cmdTribe.WithPrefix(hndlr.cfg.CommandPrefix) + " " + cmdTopODS.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopOD,
			TemplateData: map[string]interface{}{
				"Command": cmdTribe.WithPrefix(hndlr.cfg.CommandPrefix) + " " + cmdTopOD.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopPoints,
			TemplateData: map[string]interface{}{
				"Command": cmdTribe.WithPrefix(hndlr.cfg.CommandPrefix) + " " + cmdTopPoints.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpAuthor,
			TemplateData: map[string]interface{}{
				"Command": cmdAuthor.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
	)

	commandsForGuildAdmins := fmt.Sprintf(`
- %s
- %s
- %s
- %s
- %s
- %s
- %s
- %s
- %s
				`,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpAddGroup,
			TemplateData: map[string]interface{}{
				"Command": cmdAddGroup.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			TemplateData: map[string]interface{}{
				"Command": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteGroup,
			TemplateData: map[string]interface{}{
				"Command":       cmdDeleteGroup.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowEnnobledBarbs,
			TemplateData: map[string]interface{}{
				"Command":       cmdShowEnnobledBarbarians.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpObserve,
			TemplateData: map[string]interface{}{
				"Command":       cmdObserve.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			TemplateData: map[string]interface{}{
				"Command":       cmdObservations.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteObservation,
			TemplateData: map[string]interface{}{
				"Command":             cmdDeleteObservation.WithPrefix(hndlr.cfg.CommandPrefix),
				"ObservationsCommand": cmdObservations.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand":       cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpConqueredVillages,
			TemplateData: map[string]interface{}{
				"Command":       cmdConqueredVillages.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableConqueredVillages,
			TemplateData: map[string]interface{}{
				"Command":       cmdDisableConqueredVillages.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
	)

	commandsForGuildAdmins2 := fmt.Sprintf(`
- %s
- %s
- %s
- %s
- %s
- %s
				`,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpLostVillages,
			TemplateData: map[string]interface{}{
				"Command":       cmdLostVillages.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableLostVillages,
			TemplateData: map[string]interface{}{
				"Command":       cmdDisableLostVillages.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowInternals,
			TemplateData: map[string]interface{}{
				"Command":       cmdShowInternals.WithPrefix(hndlr.cfg.CommandPrefix),
				"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpChangageLanguage,
			TemplateData: map[string]interface{}{
				"Command":   cmdChangeLanguage.WithPrefix(hndlr.cfg.CommandPrefix),
				"Languages": getAvailableLanguages(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpCoordsTranslation,
			TemplateData: map[string]interface{}{
				"Command": cmdCoordsTranslation.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableCoordsTranslation,
			TemplateData: map[string]interface{}{
				"Command": cmdDisableCoordsTranslation.WithPrefix(hndlr.cfg.CommandPrefix),
			},
		}),
	)

	forAdmins := ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.HelpForAdmins,
	})

	hndlr.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTitle,
		})).
		SetURL("https://dcbot.tribalwarshelp.com/").
		SetDescription(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDescription,
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpForAllUsers,
		}), commandsForAll).
		AddField(forAdmins, commandsForGuildAdmins).
		AddField(forAdmins+" 2", commandsForGuildAdmins2))
}

type hndlrAuthor struct {
	*Session
}

var _ commandHandler = &hndlrAuthor{}

func (hndlr *hndlrAuthor) cmd() command {
	return cmdAuthor
}

func (hndlr *hndlrAuthor) requireAdmPermissions() bool {
	return false
}

func (hndlr *hndlrAuthor) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	hndlr.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Discord: Kichiyaki#2064 | https://dwysokinski.me/#contact.",
			m.Author.Mention()))
}

type hndlrTribe struct {
	*Session
}

var _ commandHandler = &hndlrTribe{}

func (hndlr *hndlrTribe) cmd() command {
	return cmdTribe
}

func (hndlr *hndlrTribe) requireAdmPermissions() bool {
	return false
}

func (hndlr *hndlrTribe) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength < 3 {
		return
	}

	command := command(args[0])
	server := args[1]
	page, err := strconv.Atoi(args[2])
	if err != nil || page <= 0 {
		hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeInvalidPage,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	var ids []int
	var tags []string
	for _, arg := range args[3:argsLength] {
		trimmed := strings.TrimSpace(arg)
		if trimmed == "" {
			continue
		}
		id, err := strconv.Atoi(trimmed)
		if err != nil || id <= 0 {
			tags = append(tags, trimmed)
		} else {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 && len(tags) == 0 {
		hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeNoTribeID,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	exists := true
	limit := 10
	offset := (page - 1) * limit
	filter := &twmodel.PlayerFilter{
		Exists: &exists,
		TribeFilter: &twmodel.TribeFilter{
			Or: &twmodel.TribeFilterOr{
				ID:  ids,
				Tag: tags,
			},
		},
	}
	title := ""
	sort := ""
	switch command {
	case cmdTopODA:
		filter.RankAttGTE = 1
		sort = "rankAtt ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODA,
		})
	case cmdTopODD:
		filter.RankDefGTE = 1
		sort = "rankDef ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODD,
		})
	case cmdTopODS:
		filter.RankSupGTE = 1
		sort = "rankSup ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODS,
		})
	case cmdTopOD:
		filter.RankTotalGTE = 1
		sort = "rankTotal ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByOD,
		})
	case cmdTopPoints:
		sort = "rank ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByPoints,
		})
	default:
		return
	}

	playerList, err := hndlr.cfg.API.Player.Browse(server,
		limit,
		offset,
		[]string{sort},
		filter,
		&sdk.PlayerInclude{
			Tribe: true,
		})
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ApiDefaultError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if playerList == nil || playerList.Total == 0 {
		hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTribesNotFound,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	totalPages := int(math.Ceil(float64(playerList.Total) / float64(limit)))
	if page > totalPages {
		hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeExceededMaximumNumberOfPages,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"Page":    page,
				"MaxPage": totalPages,
			},
		}))
		return
	}

	code := twmodel.VersionCodeFromServerKey(server)
	version, err := hndlr.cfg.API.Version.Read(code)
	if err != nil || version == nil {
		hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.InternalServerError,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	bldr := &MessageEmbedFieldBuilder{}
	bldr.SetName(title)
	for i, player := range playerList.Items {
		if player == nil {
			continue
		}

		rank := 0
		score := 0
		switch command {
		case cmdTopODA:
			rank = player.RankAtt
			score = player.ScoreAtt
		case cmdTopODD:
			rank = player.RankDef
			score = player.ScoreDef
		case cmdTopODS:
			rank = player.RankSup
			score = player.ScoreSup
		case cmdTopOD:
			rank = player.RankTotal
			score = player.ScoreTotal
		case cmdTopPoints:
			rank = player.Rank
			score = player.Points
		}

		tribeTag := "-"
		tribeURL := "-"
		if player.Tribe != nil {
			tribeTag = player.Tribe.Tag
			tribeURL = twurlbuilder.BuildTribeURL(server, version.Host, player.Tribe.ID)
		}

		bldr.Append(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeMessageLine,
			TemplateData: map[string]interface{}{
				"Index":      offset + i + 1,
				"PlayerName": player.Name,
				"PlayerURL":  twurlbuilder.BuildPlayerURL(server, version.Host, player.ID),
				"TribeTag":   tribeTag,
				"TribeURL":   tribeURL,
				"Rank":       rank,
				"Score":      humanize.Comma(int64(score)),
			},
		}))
	}

	hndlr.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(title).
		SetFields(bldr.ToMessageEmbedFields()).
		SetFooter(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.PaginationLabelDisplayedPage,
			TemplateData: map[string]interface{}{
				"Page":    page,
				"MaxPage": totalPages,
			},
		})))
}
