package discord

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tribalwarshelp/dcbot/message"
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
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

func (s *Session) handleAddGroupCommand(ctx commandCtx, m *discordgo.MessageCreate) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	if len(ctx.server.Groups) >= groupsPerServer {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.AddGroupGroupLimitHasBeenReached,
				DefaultMessage: message.FallbackMsg(message.AddGroupGroupLimitHasBeenReached,
					"{{.Mention}} The group limit has been reached ({{.Total}}/{{.Limit}})."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
					"Total":   len(ctx.server.Groups),
					"Limit":   groupsPerServer,
				},
			}))
		return
	}

	group := &models.Group{
		ServerID:               m.GuildID,
		ShowEnnobledBarbarians: true,
	}

	if err := s.cfg.GroupRepository.Store(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} Internal server error occurred, please try again later."),
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

func (s *Session) handleDeleteGroupCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteGroup,
				DefaultMessage: message.FallbackMsg(message.HelpDeleteGroup,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - deletes an observation group."),
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
				DefaultMessage: message.FallbackMsg(message.DeleteGroupInvalidID,
					"{{.Mention}} The group ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go s.cfg.GroupRepository.Delete(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DeleteGroupSuccess,
			DefaultMessage: message.FallbackMsg(message.DeleteGroupSuccess,
				"{{.Mention}} The group has been deleted."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleGroupsCommand(ctx commandCtx, m *discordgo.MessageCreate) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ServerID: []string{m.GuildID},
		Order:    []string{"id ASC"},
	})
	if err != nil {
		return
	}

	msg := ""
	for i, groups := range groups {

		msg += fmt.Sprintf("**%d** | %d | %s | %s | %s | %s\n", i+1,
			groups.ID,
			getEmojiForGroupsCommand(groups.ConqueredVillagesChannelID != ""),
			getEmojiForGroupsCommand(groups.LostVillagesChannelID != ""),
			getEmojiForGroupsCommand(groups.ShowEnnobledBarbarians),
			getEmojiForGroupsCommand(groups.ShowInternals),
		)
	}

	if msg == "" {
		msg = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsNoGroupsAdded,
			DefaultMessage: message.FallbackMsg(message.GroupsNoGroupsAdded,
				"On this server hasn't been added any groups."),
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
		}), msg).
		MessageEmbed)
}

func (s *Session) handleConqueredVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpConqueredVillages,
				DefaultMessage: message.FallbackMsg(message.HelpConqueredVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - changes the channel on which notifications about conquered village will show. **IMPORTANT!** Run this command on the channel you want to display these notifications."),
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

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
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
				"{{.Mention}} Channel changed successfully."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDisableConqueredVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
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

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
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

func (s *Session) handleLostVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpLostVillages,
				DefaultMessage: message.FallbackMsg(message.HelpLostVillages,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] changes the channel on which notifications about lost village will show. **IMPORTANT!** Run this command on the channel you want to display these notifications."),
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

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		log.Print(groups)
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
				"{{.Mention}} Channel changed successfully."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleDisableLostVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
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

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
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

func (s *Session) handleObserveCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 3 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[3:argsLength]...)
		return
	} else if argsLength < 3 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObserve,
				DefaultMessage: message.FallbackMsg(message.HelpObserve,
					"**{{.Command}}** [group id from {{.GroupsCommand}}] [server] [tribe id] - command adds a tribe to the observation group."),
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
	tribeID, err := strconv.Atoi(args[2])
	if err != nil || tribeID <= 0 {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidTribeID,
				DefaultMessage: message.FallbackMsg(message.ObserveInvalidTribeID,
					"{{.Mention}} The tribe ID must be a number greater than 0."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	server, err := s.cfg.API.Servers.Read(serverKey, nil)
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
	if server.Status == shared_models.ServerStatusClosed {
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

	tribe, err := s.cfg.API.Tribes.Read(server.Key, tribeID)
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

	err = s.cfg.ObservationRepository.Store(context.Background(), &models.Observation{
		Server:  server.Key,
		TribeID: tribeID,
		GroupID: groupID,
	})
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.ObserveSuccess,
		DefaultMessage: message.FallbackMsg(message.ObserveSuccess, "{{.Mention}} Added."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

func (s *Session) handleDeleteObservationCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 2 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[2:argsLength]...)
		return
	} else if argsLength < 2 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "help.deleteobservation",
				DefaultMessage: message.FallbackMsg("help.deleteobservation",
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
				MessageID: "deleteObservation.invalidGroupID",
				DefaultMessage: message.FallbackMsg("deleteObservation.invalidGroupID",
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
				MessageID: "deleteObservation.invalidTribeID",
				DefaultMessage: message.FallbackMsg("deleteObservation.invalidTribeID",
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
				MessageID:      "deleteObservation.groupNotFound",
				DefaultMessage: message.FallbackMsg("deleteObservation.groupNotFound", "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go s.cfg.ObservationRepository.Delete(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      "unObserve.success",
		DefaultMessage: message.FallbackMsg("unObserve.success", "{{.Mention}} Deleted."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

func (s *Session) handleObservationsCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "help.observations",
				DefaultMessage: message.FallbackMsg("help.observations",
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - shows a list of observed tribes by this group."),
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
				MessageID: "observations.invalidGroupID",
				DefaultMessage: message.FallbackMsg("observations.invalidGroupID",
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
				MessageID:      "observations.groupNotFound",
				DefaultMessage: message.FallbackMsg("observations.groupNotFound", "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observations, _, err := s.cfg.ObservationRepository.Fetch(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		Order:   []string{"id ASC"},
	})
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	tribeIDsByServer := make(map[string][]int)
	langTags := []shared_models.LanguageTag{}
	for _, observation := range observations {
		tribeIDsByServer[observation.Server] = append(tribeIDsByServer[observation.Server], observation.TribeID)
		currentLangTag := utils.LanguageTagFromServerKey(observation.Server)
		unique := true
		for _, langTag := range langTags {
			if langTag == currentLangTag {
				unique = false
				break
			}
		}
		if unique {
			langTags = append(langTags, currentLangTag)
		}
	}
	for server, tribeIDs := range tribeIDsByServer {
		list, err := s.cfg.API.Tribes.Browse(server, &shared_models.TribeFilter{
			ID: tribeIDs,
		})
		if err != nil {
			s.SendMessage(m.ChannelID,
				ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: message.InternalServerError,
					DefaultMessage: message.FallbackMsg(message.InternalServerError,
						"{{.Mention}} Internal server error occurred, please try again later."),
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
	langVersionList, err := s.cfg.API.LangVersions.Browse(&shared_models.LangVersionFilter{
		Tag: langTags,
	})

	msg := &EmbedMessage{}
	if len(observations) <= 0 || err != nil || langVersionList == nil || langVersionList.Items == nil {
		msg.Append("-")
	} else {
		for i, observation := range observations {
			tag := "Unknown"
			if observation.Tribe != nil {
				tag = observation.Tribe.Tag
			}
			lv := utils.FindLangVersionByTag(langVersionList.Items, utils.LanguageTagFromServerKey(observation.Server))
			tribeURL := ""
			if lv != nil {
				tribeURL = utils.FormatTribeURL(observation.Server, lv.Host, observation.TribeID)
			}
			msg.Append(fmt.Sprintf("**%d** | %d - %s - [``%s``](%s)\n", i+1, observation.ID,
				observation.Server,
				tag,
				tribeURL))
		}
	}
	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "observations.title",
			DefaultMessage: message.FallbackMsg("observations.title",
				"Observed tribes\nIndex | ID - Server - Tribe"),
		})).
		SetFields(msg.ToMessageEmbedFields()).
		MessageEmbed)
}

func (s *Session) handleShowEnnobledBarbariansCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "help.showennobledbarbs",
				DefaultMessage: message.FallbackMsg("help.showennobledbarbs",
					"**{{.Command}}** [group id from {{.GroupsCommand}}] - enables/disables notifications about ennobling barbarian villages."),
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
				MessageID: "showEnnobledBarbs.invalidGroupID",
				DefaultMessage: message.FallbackMsg("showEnnobledBarbs.invalidGroupID",
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
				MessageID:      "showEnnobledBarbs.groupNotFound",
				DefaultMessage: message.FallbackMsg("showEnnobledBarbs.groupNotFound", "{{.Mention}} Group not found."),
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
				DefaultMessage: message.FallbackMsg(message.InternalServerError,
					"{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "showEnnobledBarbs.success_1",
				DefaultMessage: message.FallbackMsg("showEnnobledBarbs.success_1",
					"{{.Mention}} Notifications about conquered barbarian villages will no longer show up."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	} else {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "showEnnobledBarbs.success_2",
				DefaultMessage: message.FallbackMsg("showEnnobledBarbs.success_2",
					"{{.Mention}} Enabled notifications about conquered barbarian villages."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	}
}

func (s *Session) handleChangeLanguageCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "help.changelanguage",
				DefaultMessage: message.FallbackMsg("help.changelanguage",
					"**{{.Command}}** [{{.Languages}}] - changes language."),
				TemplateData: map[string]interface{}{
					"Command":   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
					"Languages": getAvailableLanguages(),
				},
			}))
		return
	}

	lang := args[0]
	valid := false
	for _, langTag := range message.LanguageTags() {
		if langTag.String() == lang {
			valid = true
			break
		}
	}
	if !valid {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "changeLanguage.languageNotSupported",
				DefaultMessage: message.FallbackMsg("changeLanguage.languageNotSupported",
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
					"{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "changeLanguage.success",
			DefaultMessage: message.FallbackMsg("changeLanguage.success",
				"{{.Mention}} The language has been changed."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleShowInternalsCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "help.showinternals",
				DefaultMessage: message.FallbackMsg("help.showinternals",
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
				MessageID: "showInternals.invalidGroupID",
				DefaultMessage: message.FallbackMsg("showInternals.invalidGroupID",
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
				MessageID:      "showInternals.groupNotFound",
				DefaultMessage: message.FallbackMsg("showInternals.groupNotFound", "{{.Mention}} Group not found."),
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
					"{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "showInternals.success_1",
				DefaultMessage: message.FallbackMsg("showInternals.success_1",
					"{{.Mention}} Notifications about internals will no longer show up."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	} else {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "showInternals.success_2",
				DefaultMessage: message.FallbackMsg("showInternals.success_2",
					"{{.Mention}} Enabled notifications about internals."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	}
}
