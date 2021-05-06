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
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/utils"
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

func (s *Session) handleAddGroupCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
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

func (s *Session) handleDeleteGroupCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
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

func (s *Session) handleGroupsCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ServerID: []string{m.GuildID},
		DefaultFilter: models.DefaultFilter{
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
		}), msg).
		MessageEmbed)
}

func (s *Session) handleConqueredVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
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

func (s *Session) handleLostVillagesCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
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

	go s.cfg.ObservationRepository.Store(context.Background(), &models.Observation{
		Server:  server.Key,
		TribeID: tribe.ID,
		GroupID: groupID,
	})

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.ObserveSuccess,
		DefaultMessage: message.FallbackMsg(message.ObserveSuccess, "{{.Mention}} Added."),
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

	go s.cfg.ObservationRepository.Delete(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	s.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      message.DeleteObservationSuccess,
		DefaultMessage: message.FallbackMsg(message.DeleteObservationSuccess, "{{.Mention}} Deleted."),
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
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
	observations, _, err := s.cfg.ObservationRepository.Fetch(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		DefaultFilter: models.DefaultFilter{
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
	versionCodes := []twmodel.VersionCode{}
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

	msg := &MessageEmbed{}
	if len(observations) <= 0 || err != nil || versionList == nil || versionList.Items == nil {
		msg.Append("-")
	} else {
		for i, observation := range observations {
			tag := "Unknown"
			if observation.Tribe != nil {
				tag = observation.Tribe.Tag
			}
			version := utils.FindVersionByCode(versionList.Items, twmodel.VersionCodeFromServerKey(observation.Server))
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
		SetFields(msg.ToMessageEmbedFields()).
		MessageEmbed)
}

func (s *Session) handleShowEnnobledBarbariansCommand(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowEnnobledBarbs,
				DefaultMessage: message.FallbackMsg(message.HelpShowEnnobledBarbs,
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
				MessageID: message.ShowEnnobledBarbsInvalidGroupID,
				DefaultMessage: message.FallbackMsg(message.ShowEnnobledBarbsInvalidGroupID,
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
				MessageID:      message.ShowEnnobledBarbsGroupNotFound,
				DefaultMessage: message.FallbackMsg(message.ShowEnnobledBarbsGroupNotFound, "{{.Mention}} Group not found."),
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
				MessageID: message.ShowEnnobledBarbsSuccess1,
				DefaultMessage: message.FallbackMsg(message.ShowEnnobledBarbsSuccess1,
					"{{.Mention}} Notifications about conquered barbarian villages will no longer show up."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	} else {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsSuccess2,
				DefaultMessage: message.FallbackMsg(message.ShowEnnobledBarbsSuccess2,
					"{{.Mention}} Enabled notifications about conquered barbarian villages."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	}
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
	} else {
		s.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsSuccess2,
				DefaultMessage: message.FallbackMsg(message.ShowInternalsSuccess2,
					"{{.Mention}} Enabled notifications about internals."),
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
	}
}
