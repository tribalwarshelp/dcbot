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
	HelpCommand      Command = "help"
	TribeCommand     Command = "tribe"
	TopODACommand    Command = "topoda"
	TopODDCommand    Command = "topodd"
	TopODSCommand    Command = "topods"
	TopODCommand     Command = "topod"
	TopPointsCommand Command = "toppoints"
	AuthorCommand    Command = "author"
)

func (s *Session) handleHelpCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
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
			DefaultMessage: message.FallbackMsg(message.HelpTribeTopODA,
				"**{{.Command}}** [server] [page] [tribe id or tribe tag, you can enter more than one] - generates a player list from selected tribes ordered by ODA."),
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(s.cfg.CommandPrefix) + " " + TopODACommand,
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODD,
			DefaultMessage: message.FallbackMsg(message.HelpTribeTopODD,
				"**{{.Command}}** [server] [page] [tribe id or tribe tag, you can enter more than one] - generates a player list from selected tribes ordered by ODD."),
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(s.cfg.CommandPrefix) + " " + TopODDCommand,
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODS,
			DefaultMessage: message.FallbackMsg(message.HelpTribeTopODS,
				"**{{.Command}}** [server] [page] [tribe id or tribe tag, you can enter more than one] - generates a player list from selected tribes ordered by ODS."),
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(s.cfg.CommandPrefix) + " " + TopODSCommand,
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopOD,
			DefaultMessage: message.FallbackMsg(message.HelpTribeTopOD,
				"**{{.Command}}** [server] [page] [tribe id or tribe tag, you can enter more than one] - generates a player list from selected tribes ordered by OD."),
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(s.cfg.CommandPrefix) + " " + TopODCommand,
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopPoints,
			DefaultMessage: message.FallbackMsg(message.HelpTribeTopPoints,
				"**{{.Command}}** [server] [page] [tribe id or tribe tag, you can enter more than one] - generates a player list from selected tribes ordered by points."),
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(s.cfg.CommandPrefix) + " " + TopPointsCommand,
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpAuthor,
			DefaultMessage: message.FallbackMsg(message.HelpAuthor,
				"**{{.Command}}** - shows how to get in touch with the author."),
			TemplateData: map[string]interface{}{
				"Command": AuthorCommand.WithPrefix(s.cfg.CommandPrefix),
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
			DefaultMessage: message.FallbackMsg(message.HelpAddGroup,
				"**{{.Command}}** - adds a new observation group."),
			TemplateData: map[string]interface{}{
				"Command": AddGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			DefaultMessage: message.FallbackMsg(message.HelpGroups,
				"**{{.Command}}** - shows you a list of groups created by this server."),
			TemplateData: map[string]interface{}{
				"Command": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteGroup,
			DefaultMessage: message.FallbackMsg(message.HelpDeleteGroup,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - deletes an observation group."),
			TemplateData: map[string]interface{}{
				"Command":       DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowEnnobledBarbs,
			DefaultMessage: message.FallbackMsg(message.HelpShowEnnobledBarbs,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about ennobling barbarian villages."),
			TemplateData: map[string]interface{}{
				"Command":       ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpObserve,
			DefaultMessage: message.FallbackMsg(message.HelpObserve,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] [server] [tribe id or tribe tag] - adds a tribe to the observation group."),
			TemplateData: map[string]interface{}{
				"Command":       ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			DefaultMessage: message.FallbackMsg(message.HelpGroups,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - shows a list of monitored tribes added to this group."),
			TemplateData: map[string]interface{}{
				"Command":       ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteObservation,
			DefaultMessage: message.FallbackMsg(message.HelpDeleteObservation,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] [id from {{.ObservationsCommand}}] - removes a tribe from the observation group."),
			TemplateData: map[string]interface{}{
				"Command":             DeleteObservationCommand.WithPrefix(s.cfg.CommandPrefix),
				"ObservationsCommand": ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand":       GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpConqueredVillages,
			DefaultMessage: message.FallbackMsg(message.HelpConqueredVillages,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - sets the channel on which notifications about conquered village will be displayed. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableConqueredVillages,
			DefaultMessage: message.FallbackMsg(message.HelpDisableConqueredVillages,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - disables notifications about conquered villages."),
			TemplateData: map[string]interface{}{
				"Command":       DisableConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
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
			DefaultMessage: message.FallbackMsg(message.HelpLostVillages,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - sets the channel on which notifications about lost village will be displayed. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableLostVillages,
			DefaultMessage: message.FallbackMsg(message.HelpDisableLostVillages,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - sets the channel on which notifications about lost village will be displayed. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       DisableLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowInternals,
			DefaultMessage: message.FallbackMsg(message.HelpShowInternals,
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about in-group/in-tribe conquering."),
			TemplateData: map[string]interface{}{
				"Command":       ShowInternalsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpChangageLanguage,
			DefaultMessage: message.FallbackMsg(message.HelpChangageLanguage,
				"**{{.Command}}** [{{.Languages}}] - changes language."),
			TemplateData: map[string]interface{}{
				"Command":   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
				"Languages": getAvailableLanguages(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpCoordsTranslation,
			DefaultMessage: message.FallbackMsg(message.HelpCoordsTranslation,
				"**{{.Command}}** [server] - enables coords translation feature."),
			TemplateData: map[string]interface{}{
				"Command": CoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableCoordsTranslation,
			DefaultMessage: message.FallbackMsg(message.HelpDisableCoordsTranslation,
				"**{{.Command}}** - disables coords translation feature."),
			TemplateData: map[string]interface{}{
				"Command": DisableCoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
	)

	forAdmins := ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.HelpForAdmins,
		DefaultMessage: message.FallbackMsg(message.HelpForAdmins, "For admins"),
	})

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.HelpTitle,
			DefaultMessage: message.FallbackMsg(message.HelpTitle, "Help"),
		})).
		SetURL("https://dcbot.tribalwarshelp.com/").
		SetDescription(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.HelpDescription,
			DefaultMessage: message.FallbackMsg(message.HelpDescription, "Command list"),
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.HelpForAllUsers,
			DefaultMessage: message.FallbackMsg(message.HelpForAllUsers, "For everyone"),
		}), commandsForAll).
		AddField(forAdmins, commandsForGuildAdmins).
		AddField(forAdmins+" 2", commandsForGuildAdmins2).
		MessageEmbed)
}

func (s *Session) handleAuthorCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Discord: Kichiyaki#2064 | https://dwysokinski.me/#contact.",
			m.Author.Mention()))
}

func (s *Session) handleTribeCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength < 3 {
		return
	}

	command := Command(args[0])
	server := args[1]
	page, err := strconv.Atoi(args[2])
	if err != nil || page <= 0 {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeInvalidPage,
			DefaultMessage: message.FallbackMsg(message.TribeInvalidPage, "{{.Mention}} The page must be a number greater than 0."),
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
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeNoTribeID,
			DefaultMessage: message.FallbackMsg(message.TribeNoTribeID, "{{.Mention}} At least one tribe id/tag is required."),
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
	case TopODACommand:
		filter.RankAttGTE = 1
		sort = "rankAtt ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTitleOrderedByODA,
			DefaultMessage: message.FallbackMsg(message.TribeTitleOrderedByODA, "Ordered by ODA"),
		})
	case TopODDCommand:
		filter.RankDefGTE = 1
		sort = "rankDef ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTitleOrderedByODD,
			DefaultMessage: message.FallbackMsg(message.TribeTitleOrderedByODD, "Ordered by ODD"),
		})
	case TopODSCommand:
		filter.RankSupGTE = 1
		sort = "rankSup ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTitleOrderedByODS,
			DefaultMessage: message.FallbackMsg(message.TribeTitleOrderedByODS, "Ordered by ODS"),
		})
	case TopODCommand:
		filter.RankTotalGTE = 1
		sort = "rankTotal ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTitleOrderedByOD,
			DefaultMessage: message.FallbackMsg(message.TribeTitleOrderedByOD, "Ordered by OD"),
		})
	case TopPointsCommand:
		sort = "rank ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTitleOrderedByPoints,
			DefaultMessage: message.FallbackMsg(message.TribeTitleOrderedByPoints, "Ordered by points"),
		})
	default:
		return
	}

	playerList, err := s.cfg.API.Player.Browse(server,
		limit,
		offset,
		[]string{sort},
		filter,
		&sdk.PlayerInclude{
			Tribe: true,
		})
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ApiDefaultError,
				DefaultMessage: message.FallbackMsg(message.ApiDefaultError,
					"{{.Mention}} Can't fetch data from the API at the moment, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if playerList == nil || playerList.Total == 0 {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.TribeTribesNotFound,
			DefaultMessage: message.FallbackMsg(message.TribeTribesNotFound, "{{.Mention}} Tribes not found."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	totalPages := int(math.Ceil(float64(playerList.Total) / float64(limit)))
	if page > totalPages {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeExceededMaximumNumberOfPages,
			DefaultMessage: message.FallbackMsg(message.TribeExceededMaximumNumberOfPages,
				"{{.Mention}} You have exceeded the maximum number of pages ({{.Page}}/{{.MaxPage}})."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"Page":    page,
				"MaxPage": totalPages,
			},
		}))
		return
	}

	code := twmodel.VersionCodeFromServerKey(server)
	version, err := s.cfg.API.Version.Read(code)
	if err != nil || version == nil {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.InternalServerError,
			DefaultMessage: message.FallbackMsg(message.InternalServerError,
				"{{.Mention}} An internal server error has occurred, please try again later."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	msg := &MessageEmbed{}
	for i, player := range playerList.Items {
		if player == nil {
			continue
		}

		rank := 0
		score := 0
		switch command {
		case TopODACommand:
			rank = player.RankAtt
			score = player.ScoreAtt
		case TopODDCommand:
			rank = player.RankDef
			score = player.ScoreDef
		case TopODSCommand:
			rank = player.RankSup
			score = player.ScoreSup
		case TopODCommand:
			rank = player.RankTotal
			score = player.ScoreTotal
		case TopPointsCommand:
			rank = player.Rank
			score = player.Points
		}

		tribeTag := "-"
		tribeURL := "-"
		if player.Tribe != nil {
			tribeTag = player.Tribe.Tag
			tribeURL = twurlbuilder.BuildTribeURL(server, version.Host, player.Tribe.ID)
		}

		msg.Append(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeMessageLine,
			DefaultMessage: message.FallbackMsg(message.TribeMessageLine,
				"**{{.Index}}**. [`{{.PlayerName}}`]({{.PlayerURL}}) (Tribe: [`{{.TribeTag}}`]({{.TribeURL}}) | Rank: **{{.Rank}}** | Score: **{{.Score}}**)\n"),
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

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(title).
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.PaginationLabelDisplayedPage,
			DefaultMessage: message.FallbackMsg(message.PaginationLabelDisplayedPage, "{{.Page}} of {{.MaxPage}}"),
			TemplateData: map[string]interface{}{
				"Page":    page,
				"MaxPage": totalPages,
			},
		})).
		MessageEmbed)
}
