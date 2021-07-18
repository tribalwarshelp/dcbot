package discord

import (
	"context"
	"fmt"
	"github.com/tribalwarshelp/shared/tw/twmodel"
	"github.com/tribalwarshelp/shared/tw/twurlbuilder"
	"strconv"
	"strings"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/util/twutil"
)

const (
	observationsPerGroup = 10
	groupsPerServer      = 5
)

const (
	AddGroupCommand                 Command = "addgroup"
	DeleteGroupCommand              Command = "deletegroup"
	GroupsCommand                   Command = "groups"
	ShowEnnobledBarbariansCommand   Command = "showennobledbarbs"
	ObserveCommand                  Command = "observe"
	ObservationsCommand             Command = "observations"
	DeleteObservationCommand        Command = "deleteobservation"
	LostVillagesCommand             Command = "lostvillages"
	DisableLostVillagesCommand      Command = "disablelostvillages"
	ConqueredVillagesCommand        Command = "conqueredvillages"
	DisableConqueredVillagesCommand Command = "disableconqueredvillages"
	ChangeLanguageCommand           Command = "changelanguage"
	ShowInternalsCommand            Command = "showinternals"
)

type commandAddGroup struct {
	*Session
}

var _ commandHandlerInterface = &commandAddGroup{}

func (c *commandAddGroup) cmd() Command {
	return AddGroupCommand
}

func (c *commandAddGroup) requireAdmPermissions() bool {
	return true
}

func (c *commandAddGroup) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	if len(ctx.server.Groups) >= groupsPerServer {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.AddGroupLimitHasBeenReached,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
					"Total":   len(ctx.server.Groups),
					"Limit":   groupsPerServer,
				},
			}))
		return
	}

	group := &model.Group{
		ServerID:               m.GuildID,
		ShowEnnobledBarbarians: true,
	}

	if err := c.cfg.GroupRepository.Store(context.Background(), group); err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.AddGroupSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"ID":      group.ID,
			},
		}))
}

func (s *Session) handleAddGroupCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	if len(ctx.server.Groups) >= groupsPerServer {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.AddGroupLimitHasBeenReached,
				DefaultMessage: message.FallbackMsg(message.AddGroupLimitHasBeenReached,
					"{{.Mention}} The group limit has been reached ({{.Total}}/{{.Limit}})."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
					"Total":   len(ctx.server.Groups),
					"Limit":   groupsPerServer,
				},
			}))
		return
	}

	group := &model.Group{
		ServerID:               m.GuildID,
		ShowEnnobledBarbarians: true,
	}

	if err := s.cfg.GroupRepository.Store(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} An internal server error has occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.AddGroupSuccess,
			DefaultMessage: message.FallbackMsg(message.AddGroupSuccess,
				"{{.Mention}} A new group has been created (ID: {{.ID}})."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"ID":      group.ID,
			},
		}))
}

type commandDeleteGroup struct {
	*Session
}

var _ commandHandlerInterface = &commandDeleteGroup{}

func (c *commandDeleteGroup) cmd() Command {
	return DeleteGroupCommand
}

func (c *commandDeleteGroup) requireAdmPermissions() bool {
	return true
}

func (c *commandDeleteGroup) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteGroup,
				TemplateData: map[string]interface{}{
					"Command":       DeleteGroupCommand.WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteGroupInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go c.cfg.GroupRepository.Delete(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DeleteGroupSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDeleteGroupCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteGroup,
				TemplateData: map[string]interface{}{
					"Command":       DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteGroupInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go s.cfg.GroupRepository.Delete(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DeleteGroupSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandGroups struct {
	*Session
}

var _ commandHandlerInterface = &commandGroups{}

func (c *commandGroups) cmd() Command {
	return GroupsCommand
}

func (c *commandGroups) requireAdmPermissions() bool {
	return true
}

func (c *commandGroups) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	groups, _, err := c.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ServerID: []string{m.GuildID},
		DefaultFilter: model.DefaultFilter{
			Order: []string{"id ASC"},
		},
	})
	if err != nil {
		return
	}

	msg := ""
	for i, groups := range groups {
		msg += fmt.Sprintf("**%d** | %d | %s | %s | %s | %s\n", i+1,
			groups.ID,
			boolToEmoji(groups.ConqueredVillagesChannelID != ""),
			boolToEmoji(groups.LostVillagesChannelID != ""),
			boolToEmoji(groups.ShowEnnobledBarbarians),
			boolToEmoji(groups.ShowInternals),
		)
	}

	if msg == "" {
		msg = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsNoGroupsAdded,
		})
	}

	c.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsTitle,
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsFieldTitle,
		}), msg))
}

func (s *Session) handleGroupsCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ServerID: []string{m.GuildID},
		DefaultFilter: model.DefaultFilter{
			Order: []string{"id ASC"},
		},
	})
	if err != nil {
		return
	}

	msg := ""
	for i, groups := range groups {
		msg += fmt.Sprintf("**%d** | %d | %s | %s | %s | %s\n", i+1,
			groups.ID,
			boolToEmoji(groups.ConqueredVillagesChannelID != ""),
			boolToEmoji(groups.LostVillagesChannelID != ""),
			boolToEmoji(groups.ShowEnnobledBarbarians),
			boolToEmoji(groups.ShowInternals),
		)
	}

	if msg == "" {
		msg = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsNoGroupsAdded,
			DefaultMessage: message.FallbackMsg(message.GroupsNoGroupsAdded,
				"No records to display."),
		})
	}

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.GroupsTitle,
			DefaultMessage: message.FallbackMsg(message.GroupsTitle, "Group list"),
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      message.GroupsFieldTitle,
			DefaultMessage: message.FallbackMsg(message.GroupsFieldTitle, "Index | ID | Conquer | Loss | Barbarian | Internal"),
		}), msg))
}

type commandConqueredVillages struct {
	*Session
}

var _ commandHandlerInterface = &commandConqueredVillages{}

func (c *commandConqueredVillages) cmd() Command {
	return ConqueredVillagesCommand
}

func (c *commandConqueredVillages) requireAdmPermissions() bool {
	return true
}

func (c *commandConqueredVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpConqueredVillages,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := c.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].ConqueredVillagesChannelID = m.ChannelID
	go c.cfg.GroupRepository.Update(context.Background(), groups[0])
	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ConqueredVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleConqueredVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpConqueredVillages,
				DefaultMessage: message.FallbackMsg(message.HelpConqueredVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - sets the channel on which notifications about conquered village will be displayed. **IMPORTANT!** Run this command on the channel you want to display these notifications."),
				TemplateData: map[string]interface{}{
					"Command":       ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesInvalidID,
				DefaultMessage: message.FallbackMsg(message.ConqueredVillagesInvalidID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.ConqueredVillagesGroupNotFound,
					"{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].ConqueredVillagesChannelID = m.ChannelID
	go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ConqueredVillagesSuccess,
			DefaultMessage: message.FallbackMsg(message.ConqueredVillagesSuccess,
				"{{.Mention}} The channel has been successfully set."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandDisableConqueredVillages struct {
	*Session
}

var _ commandHandlerInterface = &commandDisableConqueredVillages{}

func (c *commandDisableConqueredVillages) cmd() Command {
	return DisableConqueredVillagesCommand
}

func (c *commandDisableConqueredVillages) requireAdmPermissions() bool {
	return true
}

func (c *commandDisableConqueredVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableConqueredVillages,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableConqueredVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := c.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableConqueredVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if groups[0].ConqueredVillagesChannelID != "" {
		groups[0].ConqueredVillagesChannelID = ""
		go c.cfg.GroupRepository.Update(context.Background(), groups[0])
	}
	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableConqueredVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDisableConqueredVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableConqueredVillages,
				DefaultMessage: message.FallbackMsg(message.HelpDisableConqueredVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - disables notifications about conquered villages."),
				TemplateData: map[string]interface{}{
					"Command":       DisableConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableConqueredVillagesInvalidID,
				DefaultMessage: message.FallbackMsg(message.DisableConqueredVillagesInvalidID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableConqueredVillagesGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.DisableConqueredVillagesGroupNotFound,
					"{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if groups[0].ConqueredVillagesChannelID != "" {
		groups[0].ConqueredVillagesChannelID = ""
		go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	}
	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableConqueredVillagesSuccess,
			DefaultMessage: message.FallbackMsg(message.DisableConqueredVillagesSuccess,
				"{{.Mention}} Notifications about conquered villages will no longer show up."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandLostVillages struct {
	*Session
}

var _ commandHandlerInterface = &commandLostVillages{}

func (c *commandLostVillages) cmd() Command {
	return LostVillagesCommand
}

func (c *commandLostVillages) requireAdmPermissions() bool {
	return true
}

func (c *commandLostVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpLostVillages,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := c.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].LostVillagesChannelID = m.ChannelID
	go c.cfg.GroupRepository.Update(context.Background(), groups[0])

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.LostVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleLostVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpLostVillages,
				DefaultMessage: message.FallbackMsg(message.HelpLostVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] sets the channel on which notifications about lost village will be displayed. **IMPORTANT!** Run this command on the channel you want to display these notifications."),
				TemplateData: map[string]interface{}{
					"Command":       LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesInvalidID,
				DefaultMessage: message.FallbackMsg(message.LostVillagesInvalidID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.LostVillagesGroupNotFound,
					"{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].LostVillagesChannelID = m.ChannelID
	go s.cfg.GroupRepository.Update(context.Background(), groups[0])

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.LostVillagesSuccess,
			DefaultMessage: message.FallbackMsg(message.LostVillagesSuccess,
				"{{.Mention}} The channel has been successfully set."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandDisableLostVillages struct {
	*Session
}

var _ commandHandlerInterface = &commandDisableLostVillages{}

func (c *commandDisableLostVillages) cmd() Command {
	return DisableLostVillagesCommand
}

func (c *commandDisableLostVillages) requireAdmPermissions() bool {
	return true
}

func (c *commandDisableLostVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableLostVillages,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableLostVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := c.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableLostVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if groups[0].LostVillagesChannelID != "" {
		groups[0].LostVillagesChannelID = ""
		go c.cfg.GroupRepository.Update(context.Background(), groups[0])
	}

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableLostVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDisableLostVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableLostVillages,
				DefaultMessage: message.FallbackMsg(message.HelpDisableLostVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - disables notifications about lost villages."),
				TemplateData: map[string]interface{}{
					"Command":       DisableLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableLostVillagesInvalidID,
				DefaultMessage: message.FallbackMsg(message.DisableLostVillagesInvalidID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableLostVillagesGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.DisableLostVillagesGroupNotFound,
					"{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if groups[0].LostVillagesChannelID != "" {
		groups[0].LostVillagesChannelID = ""
		go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	}

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableLostVillagesSuccess,
			DefaultMessage: message.FallbackMsg(message.DisableLostVillagesSuccess,
				"{{.Mention}} Notifications about lost villages will no longer show up."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandObserve struct {
	*Session
}

var _ commandHandlerInterface = &commandObserve{}

func (c *commandObserve) cmd() Command {
	return ObserveCommand
}

func (c *commandObserve) requireAdmPermissions() bool {
	return true
}

func (c *commandObserve) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 3 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObserve,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	serverKey := args[1]
	tribeTag := strings.TrimSpace(args[2])
	tribeID, err := strconv.Atoi(tribeTag)
	if (err != nil || tribeID <= 0) && tribeTag == "" {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidTribeID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	server, err := c.cfg.API.Server.Read(serverKey, nil)
	if err != nil || server == nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveServerNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if server.Status == twmodel.ServerStatusClosed {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveServerIsClosed,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	var tribe *twmodel.Tribe
	if tribeID > 0 {
		tribe, err = c.cfg.API.Tribe.Read(server.Key, tribeID)
	} else {
		list := &sdk.TribeList{}
		list, err = c.cfg.API.Tribe.Browse(server.Key, 1, 0, []string{}, &twmodel.TribeFilter{
			Tag: []string{tribeTag},
		})
		if list != nil && list.Items != nil && len(list.Items) > 0 {
			tribe = list.Items[0]
		}
	}
	if err != nil || tribe == nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveTribeNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := c.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if len(group.Observations) >= observationsPerGroup {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveLimitHasBeenReached,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
					"Total":   len(group.Observations),
					"Limit":   observationsPerGroup,
				},
			}))
		return
	}

	go c.cfg.ObservationRepository.Store(context.Background(), &model.Observation{
		Server:  server.Key,
		TribeID: tribe.ID,
		GroupID: groupID,
	})

	c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.ObserveSuccess,
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

func (s *Session) handleObserveCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 3 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObserve,
				DefaultMessage: message.FallbackMsg(message.HelpObserve,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] [server] [tribe id or tribe tag] - command adds a tribe to the observation group."),
				TemplateData: map[string]interface{}{
					"Command":       ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidGroupID,
				DefaultMessage: message.FallbackMsg(message.ObserveInvalidGroupID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	serverKey := args[1]
	tribeTag := strings.TrimSpace(args[2])
	tribeID, err := strconv.Atoi(tribeTag)
	if (err != nil || tribeID <= 0) && tribeTag == "" {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidTribeID,
				DefaultMessage: message.FallbackMsg(message.ObserveInvalidTribeID,
					"{{.Mention}} The third parameter must be a number greater than 0 or a valid string."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	server, err := s.cfg.API.Server.Read(serverKey, nil)
	if err != nil || server == nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ObserveServerNotFound,
				DefaultMessage: message.FallbackMsg(message.ObserveServerNotFound, "{{.Mention}} Server not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if server.Status == twmodel.ServerStatusClosed {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ObserveServerIsClosed,
				DefaultMessage: message.FallbackMsg(message.ObserveServerIsClosed, "{{.Mention}} Server is closed."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	var tribe *twmodel.Tribe
	if tribeID > 0 {
		tribe, err = s.cfg.API.Tribe.Read(server.Key, tribeID)
	} else {
		list := &sdk.TribeList{}
		list, err = s.cfg.API.Tribe.Browse(server.Key, 1, 0, []string{}, &twmodel.TribeFilter{
			Tag: []string{tribeTag},
		})
		if list != nil && list.Items != nil && len(list.Items) > 0 {
			tribe = list.Items[0]
		}
	}
	if err != nil || tribe == nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ObserveTribeNotFound,
				DefaultMessage: message.FallbackMsg(message.ObserveTribeNotFound, "{{.Mention}} Tribe not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ObserveGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.ObserveGroupNotFound, "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if len(group.Observations) >= observationsPerGroup {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveLimitHasBeenReached,
				DefaultMessage: message.FallbackMsg(message.ObserveLimitHasBeenReached,
					"{{.Mention}} The observation limit for this group has been reached ({{.Total}}/{{.Limit}})."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
					"Total":   len(group.Observations),
					"Limit":   observationsPerGroup,
				},
			}))
		return
	}

	go s.cfg.ObservationRepository.Store(context.Background(), &model.Observation{
		Server:  server.Key,
		TribeID: tribe.ID,
		GroupID: groupID,
	})

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.ObserveSuccess,
		DefaultMessage: message.FallbackMsg(message.ObserveSuccess, "{{.Mention}} The tribe has been added to the group."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

type commandDeleteObservation struct {
	*Session
}

var _ commandHandlerInterface = &commandDeleteObservation{}

func (c *commandDeleteObservation) cmd() Command {
	return DeleteObservationCommand
}

func (c *commandDeleteObservation) requireAdmPermissions() bool {
	return true
}

func (c *commandDeleteObservation) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 2 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteObservation,
				TemplateData: map[string]interface{}{
					"Command":             c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"ObservationsCommand": ObservationsCommand.WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand":       GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observationID, err := strconv.Atoi(args[1])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationInvalidTribeID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := c.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go c.cfg.ObservationRepository.Delete(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	c.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.DeleteObservationSuccess,
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

func (s *Session) handleDeleteObservationCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 2 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteObservation,
				DefaultMessage: message.FallbackMsg(message.HelpDeleteObservation,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] [id from {{.ObservationsCommand}}] - removes a tribe from the observation group."),
				TemplateData: map[string]interface{}{
					"Command":             DeleteObservationCommand.WithPrefix(s.cfg.CommandPrefix),
					"ObservationsCommand": ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand":       GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationInvalidGroupID,
				DefaultMessage: message.FallbackMsg(message.DeleteObservationInvalidGroupID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observationID, err := strconv.Atoi(args[1])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationInvalidTribeID,
				DefaultMessage: message.FallbackMsg(message.DeleteObservationInvalidTribeID,
					"{{.Mention}} The tribe ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.DeleteObservationGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.DeleteObservationGroupNotFound, "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go s.cfg.ObservationRepository.Delete(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.DeleteObservationSuccess,
		DefaultMessage: message.FallbackMsg(message.DeleteObservationSuccess, "{{.Mention}} The tribe has been removed from the group."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

type commandObservations struct {
	*Session
}

var _ commandHandlerInterface = &commandObservations{}

func (c *commandObservations) cmd() Command {
	return ObservationsCommand
}

func (c *commandObservations) requireAdmPermissions() bool {
	return true
}

func (c *commandObservations) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObservations,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObservationsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := c.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObservationsGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observations, _, err := c.cfg.ObservationRepository.Fetch(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		DefaultFilter: model.DefaultFilter{
			Order: []string{"id ASC"},
		},
	})
	if err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	tribeIDsByServer := make(map[string][]int)
	var versionCodes []twmodel.VersionCode
	for _, observation := range observations {
		tribeIDsByServer[observation.Server] = append(tribeIDsByServer[observation.Server], observation.TribeID)
		currentCode := twmodel.VersionCodeFromServerKey(observation.Server)
		unique := true
		for _, code := range versionCodes {
			if code == currentCode {
				unique = false
				break
			}
		}
		if unique {
			versionCodes = append(versionCodes, currentCode)
		}
	}
	for server, tribeIDs := range tribeIDsByServer {
		list, err := c.cfg.API.Tribe.Browse(server, 0, 0, []string{}, &twmodel.TribeFilter{
			ID: tribeIDs,
		})
		if err != nil {
			c.SendMessage(m.ChannelID,
				ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: message.InternalServerError,
					TemplateData: map[string]interface{}{
						"Mention": m.Author.Mention(),
					},
				}))
			return
		}
		for _, tribe := range list.Items {
			for _, observation := range observations {
				if observation.TribeID == tribe.ID && observation.Server == server {
					observation.Tribe = tribe
					break
				}
			}
		}
	}
	versionList, err := c.cfg.API.Version.Browse(0, 0, []string{}, &twmodel.VersionFilter{
		Code: versionCodes,
	})

	bldr := &MessageEmbedFieldBuilder{}
	if len(observations) <= 0 || err != nil || versionList == nil || versionList.Items == nil {
		bldr.Append("-")
	} else {
		for i, observation := range observations {
			tag := "Unknown"
			if observation.Tribe != nil {
				tag = observation.Tribe.Tag
			}
			version := twutil.FindVersionByCode(versionList.Items, twmodel.VersionCodeFromServerKey(observation.Server))
			tribeURL := ""
			if version != nil {
				tribeURL = twurlbuilder.BuildTribeURL(observation.Server, version.Host, observation.TribeID)
			}
			bldr.Append(fmt.Sprintf("**%d** | %d - %s - %s\n", i+1, observation.ID,
				observation.Server,
				BuildLink(tag, tribeURL)))
		}
	}
	c.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ObservationsTitle,
			DefaultMessage: message.FallbackMsg(message.ObservationsTitle,
				"Observed tribes\nIndex | ID - Server - Tribe"),
		})).
		SetFields(bldr.ToMessageEmbedFields()))
}

func (s *Session) handleObservationsCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObservations,
				DefaultMessage: message.FallbackMsg(message.HelpObservations,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - shows a list of monitored tribes added to this group."),
				TemplateData: map[string]interface{}{
					"Command":       ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObservationsInvalidGroupID,
				DefaultMessage: message.FallbackMsg(message.ObservationsInvalidGroupID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ObservationsGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.ObservationsGroupNotFound, "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observations, _, err := s.cfg.ObservationRepository.Fetch(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		DefaultFilter: model.DefaultFilter{
			Order: []string{"id ASC"},
		},
	})
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} An internal server error has occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	tribeIDsByServer := make(map[string][]int)
	var versionCodes []twmodel.VersionCode
	for _, observation := range observations {
		tribeIDsByServer[observation.Server] = append(tribeIDsByServer[observation.Server], observation.TribeID)
		currentCode := twmodel.VersionCodeFromServerKey(observation.Server)
		unique := true
		for _, code := range versionCodes {
			if code == currentCode {
				unique = false
				break
			}
		}
		if unique {
			versionCodes = append(versionCodes, currentCode)
		}
	}
	for server, tribeIDs := range tribeIDsByServer {
		list, err := s.cfg.API.Tribe.Browse(server, 0, 0, []string{}, &twmodel.TribeFilter{
			ID: tribeIDs,
		})
		if err != nil {
			s.SendMessage(m.ChannelID,
				ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: message.InternalServerError,
					DefaultMessage: message.FallbackMsg(message.InternalServerError,
						"{{.Mention}} An internal server error has occurred, please try again later."),
					TemplateData: map[string]interface{}{
						"Mention": m.Author.Mention(),
					},
				}))
			return
		}
		for _, tribe := range list.Items {
			for _, observation := range observations {
				if observation.TribeID == tribe.ID && observation.Server == server {
					observation.Tribe = tribe
					break
				}
			}
		}
	}
	versionList, err := s.cfg.API.Version.Browse(0, 0, []string{}, &twmodel.VersionFilter{
		Code: versionCodes,
	})

	msg := &MessageEmbedFieldBuilder{}
	if len(observations) <= 0 || err != nil || versionList == nil || versionList.Items == nil {
		msg.Append("-")
	} else {
		for i, observation := range observations {
			tag := "Unknown"
			if observation.Tribe != nil {
				tag = observation.Tribe.Tag
			}
			version := twutil.FindVersionByCode(versionList.Items, twmodel.VersionCodeFromServerKey(observation.Server))
			tribeURL := ""
			if version != nil {
				tribeURL = twurlbuilder.BuildTribeURL(observation.Server, version.Host, observation.TribeID)
			}
			msg.Append(fmt.Sprintf("**%d** | %d - %s - %s\n", i+1, observation.ID,
				observation.Server,
				BuildLink(tag, tribeURL)))
		}
	}
	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ObservationsTitle,
			DefaultMessage: message.FallbackMsg(message.ObservationsTitle,
				"Observed tribes\nIndex | ID - Server - Tribe"),
		})).
		SetFields(msg.ToMessageEmbedFields()))
}

type commandShowEnnobledBarbarians struct {
	*Session
}

var _ commandHandlerInterface = &commandShowEnnobledBarbarians{}

func (c *commandShowEnnobledBarbarians) cmd() Command {
	return ShowEnnobledBarbariansCommand
}

func (c *commandShowEnnobledBarbarians) requireAdmPermissions() bool {
	return true
}

func (c *commandShowEnnobledBarbarians) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowEnnobledBarbs,
				TemplateData: map[string]interface{}{
					"Command":       ShowEnnobledBarbariansCommand.WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := c.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	oldValue := group.ShowEnnobledBarbarians
	group.ShowEnnobledBarbarians = !oldValue
	if err := c.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsSuccess1,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ShowEnnobledBarbsSuccess2,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleShowEnnobledBarbariansCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowEnnobledBarbs,
				TemplateData: map[string]interface{}{
					"Command":       ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	oldValue := group.ShowEnnobledBarbarians
	group.ShowEnnobledBarbarians = !oldValue
	if err := s.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsSuccess1,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	} else {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsSuccess2,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	}
}

type commandChangeLanguage struct {
	*Session
}

var _ commandHandlerInterface = &commandChangeLanguage{}

func (c *commandChangeLanguage) cmd() Command {
	return ChangeLanguageCommand
}

func (c *commandChangeLanguage) requireAdmPermissions() bool {
	return true
}

func (c *commandChangeLanguage) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpChangageLanguage,
				TemplateData: map[string]interface{}{
					"Command":   c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"Languages": getAvailableLanguages(),
				},
			}))
		return
	}

	lang := args[0]
	valid := isValidLanguageTag(lang)
	if !valid {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ChangeLanguageLanguageNotSupported,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	ctx.server.Lang = lang
	if err := c.cfg.ServerRepository.Update(context.Background(), ctx.server); err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	ctx.localizer = message.NewLocalizer(lang)

	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ChangeLanguageSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleChangeLanguageCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpChangageLanguage,
				DefaultMessage: message.FallbackMsg(message.HelpChangageLanguage,
					"**{{.Command}}** [{{.Languages}}] - changes language."),
				TemplateData: map[string]interface{}{
					"Command":   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
					"Languages": getAvailableLanguages(),
				},
			}))
		return
	}

	lang := args[0]
	valid := isValidLanguageTag(lang)
	if !valid {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ChangeLanguageLanguageNotSupported,
				DefaultMessage: message.FallbackMsg(message.ChangeLanguageLanguageNotSupported,
					"{{.Mention}} Language not supported."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	ctx.server.Lang = lang
	if err := s.cfg.ServerRepository.Update(context.Background(), ctx.server); err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} An internal server error has occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	ctx.localizer = message.NewLocalizer(lang)

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ChangeLanguageSuccess,
			DefaultMessage: message.FallbackMsg(message.ChangeLanguageSuccess,
				"{{.Mention}} The language has been changed."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type commandShowInternals struct {
	*Session
}

var _ commandHandlerInterface = &commandShowInternals{}

func (c *commandShowInternals) cmd() Command {
	return ShowInternalsCommand
}

func (c *commandShowInternals) requireAdmPermissions() bool {
	return true
}

func (c *commandShowInternals) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		c.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowInternals,
				TemplateData: map[string]interface{}{
					"Command":       c.cmd().WithPrefix(c.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(c.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := c.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	oldValue := group.ShowInternals
	group.ShowInternals = !oldValue
	if err := c.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		c.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsSuccess1,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	c.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ShowInternalsSuccess2,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleShowInternalsCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowInternals,
				DefaultMessage: message.FallbackMsg(message.HelpShowInternals,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about in-group/in-tribe conquering."),
				TemplateData: map[string]interface{}{
					"Command":       ShowInternalsCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsInvalidGroupID,
				DefaultMessage: message.FallbackMsg(message.ShowInternalsInvalidGroupID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      message.ShowInternalsGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.ShowInternalsGroupNotFound, "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	oldValue := group.ShowInternals
	group.ShowInternals = !oldValue
	if err := s.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} An internal server error has occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsSuccess1,
				DefaultMessage: message.FallbackMsg(message.ShowInternalsSuccess1,
					"{{.Mention}} Notifications about internals will no longer show up."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ShowInternalsSuccess2,
			DefaultMessage: message.FallbackMsg(message.ShowInternalsSuccess2,
				"{{.Mention}} Notifications about internals have been enabled."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}
