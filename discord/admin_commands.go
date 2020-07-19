package discord

import (
	"context"
	"fmt"
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
	AddGroupCommand                   Command = "addgroup"
	DeleteGroupCommand                Command = "deletegroup"
	GroupsCommand                     Command = "groups"
	ShowEnnobledBarbariansCommand     Command = "showennobledbarbs"
	ObserveCommand                    Command = "observe"
	ObservationsCommand               Command = "observations"
	UnObserveCommand                  Command = "unobserve"
	LostVillagesCommand               Command = "lostvillages"
	UnObserveLostVillagesCommand      Command = "unobservelostvillages"
	ConqueredVillagesCommand          Command = "conqueredvillages"
	UnObserveConqueredVillagesCommand Command = "unobserveconqueredvillages"
)

func (s *Session) handleAddGroupCommand(ctx commandCtx, m *discordgo.MessageCreate) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	if len(ctx.server.Groups) >= groupsPerServer {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      "addGroup.groupLimitHasBeenReached",
				DefaultMessage: message.FallbackMsg("addGroup.groupLimitHasBeenReached", "{{.Mention}} The group limit has been reached ({{.Total}}/{{.Limit}})."),
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
				MessageID:      "internalServerError",
				DefaultMessage: message.FallbackMsg("internalServerError", "{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	s.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "addGroup.success",
			DefaultMessage: message.FallbackMsg("addGroup.success", "{{.Mention}} A new group has been created (ID: {{.ID}})."),
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
				MessageID:      "help.deletegroup",
				DefaultMessage: message.FallbackMsg("help.deletegroup", "**{{.Command}}** [group id from {{.GroupsCommand}}] - deletes an observation group."),
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
				MessageID:      "deleteGroup.invalidID",
				DefaultMessage: message.FallbackMsg("deleteGroup.invalidID", "{{.Mention}} The group ID must be a number greater than 0."),
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
			MessageID:      "deleteGroup.success",
			DefaultMessage: message.FallbackMsg("deleteGroup.success", "{{.Mention}} The group has been deleted."),
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
		msg += fmt.Sprintf("**%d** | %d\n", i+1, groups.ID)
	}

	if msg == "" {
		msg = ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "groups.noGroupsAdded",
			DefaultMessage: message.FallbackMsg("groups.noGroupsAdded", "No groups added"),
		})
	}

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:      "groups.title",
			DefaultMessage: message.FallbackMsg("groups.title", "Group list"),
		})).
		AddField("Index | ID", msg).
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
				MessageID:      "help.conqueredvillages",
				DefaultMessage: message.FallbackMsg("help.conqueredvillages", "**{{.Command}}** [group id from {{.GroupsCommand}}] - changes the channel on which notifications about conquered village will show. **IMPORTANT!** Call this command on the channel you want to display these notifications."),
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
				MessageID:      "conqueredVillages.invalidID",
				DefaultMessage: message.FallbackMsg("conqueredVillages.invalidID", "{{.Mention}} The group ID must be a number greater than 0."),
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
				MessageID:      "conqueredVillages.groupNotFound",
				DefaultMessage: message.FallbackMsg("conqueredVillages.groupNotFound", "{{.Mention}} Group not found."),
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
			MessageID:      "conqueredVillages.success",
			DefaultMessage: message.FallbackMsg("conqueredVillages.success", "{{.Mention}} Channel changed successfully."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleUnObserveConqueredVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
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
				MessageID:      "help.unobserveconqueredvillages",
				DefaultMessage: message.FallbackMsg("help.unobserveconqueredvillages", "**{{.Command}}** [group id from {{.GroupsCommand}}] - disable notifications about conquered villages."),
				TemplateData: map[string]interface{}{
					"Command":       UnObserveConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      "unObserveConqueredVillages.invalidID",
				DefaultMessage: message.FallbackMsg("unObserveConqueredVillages.invalidID", "{{.Mention}} The group ID must be a number greater than 0."),
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
				MessageID:      "unObserveConqueredVillages.groupNotFound",
				DefaultMessage: message.FallbackMsg("unObserveConqueredVillages.groupNotFound", "{{.Mention}} Group not found."),
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
			MessageID:      "unObserveConqueredVillages.success",
			DefaultMessage: message.FallbackMsg("unObserveConqueredVillages.success", "{{.Mention}} Notifications about conquered villages will no longer show up."),
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
				MessageID:      "help.lostvillages",
				DefaultMessage: message.FallbackMsg("help.lostvillages", "**{{.Command}}** [group id from {{.GroupsCommand}}] changes the channel on which notifications about lost village will show. **IMPORTANT!** Call this command on the channel you want to display these notifications."),
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
				MessageID:      "lostVillages.invalidID",
				DefaultMessage: message.FallbackMsg("lostVillages.invalidID", "{{.Mention}} The group ID must be a number greater than 0."),
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
				MessageID:      "lostVillages.groupNotFound",
				DefaultMessage: message.FallbackMsg("lostVillages.groupNotFound", "{{.Mention}} Group not found."),
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
			MessageID:      "lostVillages.success",
			DefaultMessage: message.FallbackMsg("lostVillages.success", "{{.Mention}} Channel changed successfully."),
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

func (s *Session) handleUnObserveLostVillagesCommand(ctx commandCtx, m *discordgo.MessageCreate, args ...string) {
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
				MessageID:      "help.unobservelostvillages",
				DefaultMessage: message.FallbackMsg("help.unobservelostvillages", "**{{.Command}}** [group id from {{.GroupsCommand}}] - disable notifications about lost villages."),
				TemplateData: map[string]interface{}{
					"Command":       UnObserveLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
					"GroupsCommand": GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      "unObserveLostVillages.invalidID",
				DefaultMessage: message.FallbackMsg("unObserveLostVillages.invalidID", "{{.Mention}} The group ID must be a number greater than 0."),
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
				MessageID:      "unObserveLostVillages.groupNotFound",
				DefaultMessage: message.FallbackMsg("unObserveLostVillages.groupNotFound", "{{.Mention}} Group not found."),
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
			MessageID:      "unObserveLostVillages.success",
			DefaultMessage: message.FallbackMsg("unObserveLostVillages.success", "{{.Mention}} Notifications about lost villages will no longer show up."),
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
				MessageID:      "help.observe",
				DefaultMessage: message.FallbackMsg("help.observe", "**{{.Command}}** [group id from {{.GroupsCommand}}] [server] [tribe id] - command adds a tribe to the observation group."),
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
				MessageID:      "observe.invalidGroupID",
				DefaultMessage: message.FallbackMsg("observe.invalidGroupID", "{{.Mention}} The group ID must be a number greater than 0."),
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
				MessageID:      "observe.invalidTribeID",
				DefaultMessage: message.FallbackMsg("observe.invalidTribeID", "{{.Mention}} The tribe ID must be a number greater than 0."),
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
				MessageID:      "observe.serverNotFound",
				DefaultMessage: message.FallbackMsg("observe.serverNotFound", "{{.Mention}} Server not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if server.Status == shared_models.ServerStatusClosed {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      "observe.serverIsClosed",
				DefaultMessage: message.FallbackMsg("observe.serverIsClosed", "{{.Mention}} Server is closed."),
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
				MessageID:      "observe.tribeNotFound",
				DefaultMessage: message.FallbackMsg("observe.tribeNotFound", "{{.Mention}} Tribe not found."),
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
				MessageID:      "observe.groupNotFound",
				DefaultMessage: message.FallbackMsg("observe.groupNotFound", "{{.Mention}} Group not found."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if len(group.Observations) >= observationsPerGroup {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:      "observe.observationLimitHasBeenReached",
				DefaultMessage: message.FallbackMsg("observe.observationLimitHasBeenReached", "{{.Mention}} The observation limit for this group has been reached ({{.Total}}/{{.Limit}})."),
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
				MessageID:      "internalServerError",
				DefaultMessage: message.FallbackMsg("internalServerError", "{{.Mention}} Internal server error occurred, please try again later."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      "observe.success",
		DefaultMessage: message.FallbackMsg("observe.success", "{{.Mention}} Added."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

func (s *Session) handleUnObserveCommand(m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 2 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[2:argsLength]...)
		return
	} else if argsLength < 2 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id grupy] [id obserwacji]`,
				m.Author.Mention(),
				UnObserveCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID grupy powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}
	observationID, err := strconv.Atoi(args[1])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID obserwacji powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}

	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}

	go s.cfg.ObservationRepository.Delete(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	s.SendMessage(m.ChannelID, m.Author.Mention()+` Usunięto.`)
}

func (s *Session) handleObservationsCommand(m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id grupy]`,
				m.Author.Mention(),
				ObservationsCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID grupy powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}
	observations, _, err := s.cfg.ObservationRepository.Fetch(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		Order:   []string{"id ASC"},
	})
	if err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Wystąpił błąd wewnętrzny, prosimy spróbować później.`)
		return
	}

	tribeIDsByServer := make(map[string][]int)
	observationIndexByTribeID := make(map[int]int)
	langTags := []shared_models.LanguageTag{}
	for i, observation := range observations {
		tribeIDsByServer[observation.Server] = append(tribeIDsByServer[observation.Server], observation.TribeID)
		observationIndexByTribeID[observation.TribeID] = i
		currentLangTag := utils.LanguageTagFromWorldName(observation.Server)
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
			s.SendMessage(m.ChannelID, m.Author.Mention()+` Wystąpił błąd wewnętrzny, prosimy spróbować później.`)
			return
		}
		for _, tribe := range list.Items {
			observations[observationIndexByTribeID[tribe.ID]].Tribe = tribe
		}
	}
	langVersionList, err := s.cfg.API.LangVersions.Browse(&shared_models.LangVersionFilter{
		Tag: langTags,
	})

	msg := &EmbedMessage{}
	if len(observations) <= 0 {
		msg.Append("Brak")
	}
	for i, observation := range observations {
		tag := "Unknown"
		if observation.Tribe != nil {
			tag = observation.Tribe.Tag
		}
		lv := utils.FindLangVersionByTag(langVersionList.Items, utils.LanguageTagFromWorldName(observation.Server))
		tribeURL := ""
		if lv != nil {
			tribeURL = utils.FormatTribeURL(observation.Server, lv.Host, observation.TribeID)
		}
		msg.Append(fmt.Sprintf("**%d** | %d - %s - [``%s``](%s)\n", i+1, observation.ID,
			observation.Server,
			tag,
			tribeURL))
	}
	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Lista obserwowanych plemion\nIndeks | ID - Serwer - Plemię").
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter("Strona 1 z 1").
		MessageEmbed)
}

func (s *Session) handleShowEnnobledBarbariansCommand(m *discordgo.MessageCreate, args ...string) {
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id grupy]`,
				m.Author.Mention(),
				ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID grupy powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}

	oldValue := group.ShowEnnobledBarbarians
	group.ShowEnnobledBarbarians = !oldValue
	if err := s.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Wewnętrzny błąd serwera.`)
		return
	}

	if oldValue {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+` Informacje o podbitych wioskach barbarzyńskich nie będą się już pojawiały.`)
	} else {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+` Włączono wyświetlanie informacji o podbitych wioskach barbarzyńskich.`)
	}
}
