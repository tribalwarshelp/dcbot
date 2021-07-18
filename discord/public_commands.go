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

type commandHelp struct {
	*Session
}

var _ commandHandler = &commandHelp{}

func (c *commandHelp) cmd() Command {
	return HelpCommand
}

func (c *commandHelp) requireAdmPermissions() bool {
	return false
}

func (c *commandHelp) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
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
				"Command": TribeCommand.WithPrefix(c.cfg.CommandPrefix) + " " + TopODACommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODD,
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(c.cfg.CommandPrefix) + " " + TopODDCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopODS,
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(c.cfg.CommandPrefix) + " " + TopODSCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopOD,
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(c.cfg.CommandPrefix) + " " + TopODCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpTribeTopPoints,
			TemplateData: map[string]interface{}{
				"Command": TribeCommand.WithPrefix(c.cfg.CommandPrefix) + " " + TopPointsCommand.String(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpAuthor,
			TemplateData: map[string]interface{}{
				"Command": AuthorCommand.WithPrefix(c.cfg.CommandPrefix),
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
				"Command": AddGroupCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			TemplateData: map[string]interface{}{
				"Command": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteGroup,
			TemplateData: map[string]interface{}{
				"Command":       DeleteGroupCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowEnnobledBarbs,
			TemplateData: map[string]interface{}{
				"Command":       ShowEnnobledBarbariansCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpObserve,
			TemplateData: map[string]interface{}{
				"Command":       ObserveCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpGroups,
			TemplateData: map[string]interface{}{
				"Command":       ObservationsCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDeleteObservation,
			TemplateData: map[string]interface{}{
				"Command":             DeleteObservationCommand.WithPrefix(c.cfg.CommandPrefix),
				"ObservationsCommand": ObservationsCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand":       GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpConqueredVillages,
			TemplateData: map[string]interface{}{
				"Command":       ConqueredVillagesCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableConqueredVillages,
			TemplateData: map[string]interface{}{
				"Command":       DisableConqueredVillagesCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
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
				"Command":       LostVillagesCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableLostVillages,
			TemplateData: map[string]interface{}{
				"Command":       DisableLostVillagesCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpShowInternals,
			TemplateData: map[string]interface{}{
				"Command":       ShowInternalsCommand.WithPrefix(c.cfg.CommandPrefix),
				"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpChangageLanguage,
			TemplateData: map[string]interface{}{
				"Command":   ChangeLanguageCommand.WithPrefix(c.cfg.CommandPrefix),
				"Languages": getAvailableLanguages(),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpCoordsTranslation,
			TemplateData: map[string]interface{}{
				"Command": CoordsTranslationCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.HelpDisableCoordsTranslation,
			TemplateData: map[string]interface{}{
				"Command": DisableCoordsTranslationCommand.WithPrefix(c.cfg.CommandPrefix),
			},
		}),
	)

	forAdmins := ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.HelpForAdmins,
	})

	c.SendEmbed(m.ChannelID, NewEmbed().
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

type commandAuthor struct {
	*Session
}

var _ commandHandler = &commandAuthor{}

func (c *commandAuthor) cmd() Command {
	return AuthorCommand
}

func (c *commandAuthor) requireAdmPermissions() bool {
	return false
}

func (c *commandAuthor) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	c.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Discord: Kichiyaki#2064 | https://dwysokinski.me/#contact.",
			m.Author.Mention()))
}

type commandTribe struct {
	*Session
}

var _ commandHandler = &commandTribe{}

func (c *commandTribe) cmd() Command {
	return TribeCommand
}

func (c *commandTribe) requireAdmPermissions() bool {
	return false
}

func (c *commandTribe) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength < 3 {
		return
	}

	command := Command(args[0])
	server := args[1]
	page, err := strconv.Atoi(args[2])
	if err != nil || page <= 0 {
		c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
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
		c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
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
	case TopODACommand:
		filter.RankAttGTE = 1
		sort = "rankAtt ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODA,
		})
	case TopODDCommand:
		filter.RankDefGTE = 1
		sort = "rankDef ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODD,
		})
	case TopODSCommand:
		filter.RankSupGTE = 1
		sort = "rankSup ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByODS,
		})
	case TopODCommand:
		filter.RankTotalGTE = 1
		sort = "rankTotal ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByOD,
		})
	case TopPointsCommand:
		sort = "rank ASC"
		title = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTitleOrderedByPoints,
		})
	default:
		return
	}

	playerList, err := c.cfg.API.Player.Browse(server,
		limit,
		offset,
		[]string{sort},
		filter,
		&sdk.PlayerInclude{
			Tribe: true,
		})
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ApiDefaultError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if playerList == nil || playerList.Total == 0 {
		c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.TribeTribesNotFound,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}
	totalPages := int(math.Ceil(float64(playerList.Total) / float64(limit)))
	if page > totalPages {
		c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
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
	version, err := c.cfg.API.Version.Read(code)
	if err != nil || version == nil {
		c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.InternalServerError,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
		return
	}

	bldr := &MessageEmbedFieldBuilder{}
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

	c.SendEmbed(m.ChannelID, NewEmbed().
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
