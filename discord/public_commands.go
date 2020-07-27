package discord

import (
	"fmt"
	"math"
	"strconv"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/bwmarrin/discordgo"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

const (
	HelpCommand      Command = "help"
	TribeCommand     Command = "tribe"
	TopAttCommand    Command = "topatt"
	TopDefCommand    Command = "topdef"
	TopSuppCommand   Command = "topsupp"
	TopTotalCommand  Command = "toptotal"
	TopPointsCommand Command = "toppoints"
	AuthorCommand    Command = "author"
)

func (s *Session) handleHelpCommand(ctx commandCtx, m *discordgo.MessageCreate) {
	tribeCMDWithPrefix := TribeCommand.WithPrefix(s.cfg.CommandPrefix)
	commandsForAll := fmt.Sprintf(`
- %s
- %s
- %s
- %s
- %s
- %s
				`,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.tribe.topatt",
			DefaultMessage: message.FallbackMsg("help.tribe.topatt",
				"**{{.Command}}** [server] [page] [id1] [id2] [id3] [n id] - generates a player list from selected tribes ordered by ODA."),
			TemplateData: map[string]interface{}{
				"Command": tribeCMDWithPrefix + " " + TopAttCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.tribe.topdef",
			DefaultMessage: message.FallbackMsg("help.tribe.topdef",
				"**{{.Command}}** [server] [page] [id1] [id2] [id3] [n id] - generates a player list from selected tribes ordered by ODD."),
			TemplateData: map[string]interface{}{
				"Command": tribeCMDWithPrefix + " " + TopDefCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.tribe.topsupp",
			DefaultMessage: message.FallbackMsg("help.tribe.topsupp",
				"**{{.Command}}** [server] [page] [id1] [id2] [id3] [n id] - generates a player list from selected tribes ordered by ODS."),
			TemplateData: map[string]interface{}{
				"Command": tribeCMDWithPrefix + " " + TopSuppCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.tribe.toptotal",
			DefaultMessage: message.FallbackMsg("help.tribe.toptotal",
				"**{{.Command}}** [server] [page] [id1] [id2] [id3] [n id] - generates a player list from selected tribes ordered by OD."),
			TemplateData: map[string]interface{}{
				"Command": tribeCMDWithPrefix + " " + TopTotalCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.tribe.toppoints",
			DefaultMessage: message.FallbackMsg("help.tribe.toppoints",
				"**{{.Command}}** [server] [page] [id1] [id2] [id3] [n id] - generates a player list from selected tribes ordered by points."),
			TemplateData: map[string]interface{}{
				"Command": tribeCMDWithPrefix + " " + TopPointsCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.author",
			DefaultMessage: message.FallbackMsg("help.author",
				"**{{.Command}}** - shows how to contact the author."),
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
			MessageID: "help.addgroup",
			DefaultMessage: message.FallbackMsg("help.addgroup",
				"**{{.Command}}** - adds a new observation group."),
			TemplateData: map[string]interface{}{
				"Command": AddGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.groups",
			DefaultMessage: message.FallbackMsg("help.groups",
				"**{{.Command}}** - shows you a list of groups created by this guild."),
			TemplateData: map[string]interface{}{
				"Command": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.deletegroup",
			DefaultMessage: message.FallbackMsg("help.deletegroup",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - deletes an observation group."),
			TemplateData: map[string]interface{}{
				"Command":       DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.showennobledbarbs",
			DefaultMessage: message.FallbackMsg("help.showennobledbarbs",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about ennobling barbarian villages."),
			TemplateData: map[string]interface{}{
				"Command":       ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.observe",
			DefaultMessage: message.FallbackMsg("help.observe",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] [server] [tribe id] - command adds a tribe to the observation group."),
			TemplateData: map[string]interface{}{
				"Command":       ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.observations",
			DefaultMessage: message.FallbackMsg("help.observations",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] shows a list of observed tribes by this group."),
			TemplateData: map[string]interface{}{
				"Command":       ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.deleteobservation",
			DefaultMessage: message.FallbackMsg("help.deleteobservation",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] [id from {{.ObservationsCommand}}] - removes a tribe from the observation group."),
			TemplateData: map[string]interface{}{
				"Command":             DeleteObservationCommand.WithPrefix(s.cfg.CommandPrefix),
				"ObservationsCommand": ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand":       GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.conqueredvillages",
			DefaultMessage: message.FallbackMsg("help.conqueredvillages",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - changes the channel on which notifications about conquered village will show. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.disableconqueredvillages",
			DefaultMessage: message.FallbackMsg("help.disableconqueredvillages",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - disable notifications about conquered villages."),
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
				`,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.lostvillages",
			DefaultMessage: message.FallbackMsg("help.lostvillages",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - changes the channel on which notifications about lost village will show. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.disablelostvillages",
			DefaultMessage: message.FallbackMsg("help.disablelostvillages",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - changes the channel on which notifications about lost village will show. IMPORTANT! Call this command on the channel you want to display these notifications."),
			TemplateData: map[string]interface{}{
				"Command":       DisableLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.changelanguage",
			DefaultMessage: message.FallbackMsg("help.changelanguage",
				"**{{.Command}}** [{{.Languages}}] - change language."),
			TemplateData: map[string]interface{}{
				"Command":   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
				"Languages": getAvailableLanguages(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "help.showinternals",
			DefaultMessage: message.FallbackMsg("help.showinternals",
				"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about self-conquers between tribes in one group."),
			TemplateData: map[string]interface{}{
				"Command":       ShowInternalsCommand.WithPrefix(s.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			},
		}),
	)

	forAdmins := ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      "help.forAdmins",
		DefaultMessage: message.FallbackMsg("help.forAdmins", "For admins"),
	})

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "help.title",
			DefaultMessage: message.FallbackMsg("help.title", "Help"),
		})).
		SetURL("https://dcbot.tribalwarshelp.com/").
		SetDescription(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "help.description",
			DefaultMessage: message.FallbackMsg("help.description", "Commands offered by bot"),
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "help.forAllUsers",
			DefaultMessage: message.FallbackMsg("help.forAllUsers", "For all guild members."),
		}), commandsForAll).
		AddField(forAdmins, commandsForGuildAdmins).
		AddField(forAdmins+" 2", commandsForGuildAdmins2).
		MessageEmbed)
}

func (s *Session) handleAuthorCommand(m *discordgo.MessageCreate) {
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Discord: Kichiyaki#2064 | https://dawid-wysokinski.pl/#contact.",
			m.Author.Mention()))
}

func (s *Session) handleTribeCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength < 3 {
		return
	}

	command := Command(args[0])
	world := args[1]
	page, err := strconv.Atoi(args[2])
	if err != nil || page <= 0 {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.invalidPage",
			DefaultMessage: message.FallbackMsg("tribe.invalidPage", "{{.Mention}} The page must be a number greater than 0."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	ids := []int{}
	for _, arg := range args[3:argsLength] {
		id, err := strconv.Atoi(arg)
		if err != nil || id <= 0 {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.noTribeID",
			DefaultMessage: message.FallbackMsg("tribe.noTribeID", "{{.Mention}} You haven't entered the tribe ID."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	exists := true
	limit := 10
	offset := (page - 1) * limit
	filter := &shared_models.PlayerFilter{
		Exists:  &exists,
		TribeID: ids,
		Limit:   limit,
		Offset:  offset,
	}
	title := ""
	switch command {
	case TopAttCommand:
		filter.RankAttGTE = 1
		filter.Sort = "rankAtt ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.title.orderedByODA",
			DefaultMessage: message.FallbackMsg("tribe.title.orderedByODA", "Ordered by ODA"),
		})
	case TopDefCommand:
		filter.RankDefGTE = 1
		filter.Sort = "rankDef ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.title.orderedByODD",
			DefaultMessage: message.FallbackMsg("tribe.title.orderedByODD", "Ordered by ODD"),
		})
	case TopSuppCommand:
		filter.RankSupGTE = 1
		filter.Sort = "rankSup ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.title.orderedByODS",
			DefaultMessage: message.FallbackMsg("tribe.title.orderedByODS", "Ordered by ODS"),
		})
	case TopTotalCommand:
		filter.RankTotalGTE = 1
		filter.Sort = "rankTotal ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.title.orderedByOD",
			DefaultMessage: message.FallbackMsg("tribe.title.orderedByOD", "Ordered by OD"),
		})
	case TopPointsCommand:
		filter.Sort = "rank ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "tribe.title.orderedByPoints",
			DefaultMessage: message.FallbackMsg("tribe.title.orderedByPoints", "Ordered by points"),
		})
	default:
		return
	}

	playersList, err := s.cfg.API.Players.Browse(world, filter, &sdk.PlayerInclude{
		Tribe: true,
	})
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "api.defaultError",
				DefaultMessage: message.FallbackMsg("api.defaultError",
					"{{.Mention}} There was an error fetching data from the API, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if playersList == nil || playersList.Total == 0 {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "tribe.tribesNotFound",
			DefaultMessage: message.FallbackMsg("tribe.tribesNotFound",
				"{{.Mention}} There was an error fetching data from the API, please try again later."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	totalPages := int(math.Ceil(float64(playersList.Total) / float64(limit)))
	if page > totalPages {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "tribe.exceededMaximumNumberOfPages",
			DefaultMessage: message.FallbackMsg("tribe.exceededMaximumNumberOfPages",
				"{{.Mention}} You have exceeded the maximum number of pages ({{.Page}}/{{.MaxPage}})."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"Page":    page,
				"MaxPage": totalPages,
			},
		}))
		return
	}

	langTag := utils.LanguageTagFromWorldName(world)
	langVersion, err := s.cfg.API.LangVersions.Read(langTag)
	if err != nil || langVersion == nil {
		s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "internalServerError",
			DefaultMessage: message.FallbackMsg("internalServerError",
				"{{.Mention}} Internal server error occurred, please try again later."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	msg := &EmbedMessage{}
	for i, player := range playersList.Items {
		if player == nil {
			continue
		}

		rank := 0
		score := 0
		switch command {
		case TopAttCommand:
			rank = player.RankAtt
			score = player.ScoreAtt
		case TopDefCommand:
			rank = player.RankDef
			score = player.ScoreDef
		case TopSuppCommand:
			rank = player.RankSup
			score = player.ScoreSup
		case TopTotalCommand:
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
			tribeURL = utils.FormatTribeURL(world, langVersion.Host, player.Tribe.ID)
		}

		msg.Append(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "tribe.messageLine",
			DefaultMessage: message.FallbackMsg("tribe.messageLine",
				"**{{.Index}}**. [``{{.PlayerName}}``]({{.PlayerURL}}) (Tribe: [``{{.TribeTag}}``]({{.TribeURL}}) | Rank: **{{.Rank}}** | Score: **{{.Score}}**)\n"),
			TemplateData: map[string]interface{}{
				"Index":      offset + i + 1,
				"PlayerName": player.Name,
				"PlayerURL":  utils.FormatPlayerURL(world, langVersion.Host, player.ID),
				"TribeTag":   tribeTag,
				"TribeURL":   tribeURL,
				"Rank":       rank,
				"Score":      score,
			},
		}))
	}

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(title).
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "pagination.labelDisplayedPage",
			DefaultMessage: message.FallbackMsg("pagination.labelDisplayedPage", "{{.Page}} of {{.MaxPage}}"),
			TemplateData: map[string]interface{}{
				"Page":    page,
				"MaxPage": totalPages,
			},
		})).
		MessageEmbed)
}
