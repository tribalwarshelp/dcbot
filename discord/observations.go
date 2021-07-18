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
	cmdAddGroup                 command = "addgroup"
	cmdDeleteGroup              command = "deletegroup"
	cmdGroups                   command = "groups"
	cmdShowEnnobledBarbarians   command = "showennobledbarbs"
	cmdObserve                  command = "observe"
	cmdObservations             command = "observations"
	cmdDeleteObservation        command = "deleteobservation"
	cmdLostVillages             command = "lostvillages"
	cmdDisableLostVillages      command = "disablelostvillages"
	cmdConqueredVillages        command = "conqueredvillages"
	cmdDisableConqueredVillages command = "disableconqueredvillages"
	cmdShowInternals            command = "showinternals"
)

type hndlrAddGroup struct {
	*Session
}

var _ commandHandler = &hndlrAddGroup{}

func (hndlr *hndlrAddGroup) cmd() command {
	return cmdAddGroup
}

func (hndlr *hndlrAddGroup) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrAddGroup) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	if len(ctx.server.Groups) >= groupsPerServer {
		hndlr.SendMessage(m.ChannelID,
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

	if err := hndlr.cfg.GroupRepository.Store(context.Background(), group); err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.AddGroupSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
				"ID":      group.ID,
			},
		}))
}

type hndlrDeleteGroup struct {
	*Session
}

var _ commandHandler = &hndlrDeleteGroup{}

func (hndlr *hndlrDeleteGroup) cmd() command {
	return cmdDeleteGroup
}

func (hndlr *hndlrDeleteGroup) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrDeleteGroup) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteGroup,
				TemplateData: map[string]interface{}{
					"Command":       cmdDeleteGroup.WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteGroupInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go hndlr.cfg.GroupRepository.Delete(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})

	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DeleteGroupSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrGroups struct {
	*Session
}

var _ commandHandler = &hndlrGroups{}

func (hndlr *hndlrGroups) cmd() command {
	return cmdGroups
}

func (hndlr *hndlrGroups) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrGroups) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	groups, _, err := hndlr.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
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

	hndlr.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsTitle,
		})).
		AddField(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.GroupsFieldTitle,
		}), msg))
}

type hndlrConqueredVillages struct {
	*Session
}

var _ commandHandler = &hndlrConqueredVillages{}

func (hndlr *hndlrConqueredVillages) cmd() command {
	return cmdConqueredVillages
}

func (hndlr *hndlrConqueredVillages) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrConqueredVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpConqueredVillages,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := hndlr.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ConqueredVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].ConqueredVillagesChannelID = m.ChannelID
	go hndlr.cfg.GroupRepository.Update(context.Background(), groups[0])
	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ConqueredVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrDisableConqueredVillages struct {
	*Session
}

var _ commandHandler = &hndlrDisableConqueredVillages{}

func (hndlr *hndlrDisableConqueredVillages) cmd() command {
	return cmdDisableConqueredVillages
}

func (hndlr *hndlrDisableConqueredVillages) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrDisableConqueredVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableConqueredVillages,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableConqueredVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := hndlr.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		hndlr.SendMessage(m.ChannelID,
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
		go hndlr.cfg.GroupRepository.Update(context.Background(), groups[0])
	}
	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableConqueredVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrLostVillages struct {
	*Session
}

var _ commandHandler = &hndlrLostVillages{}

func (hndlr *hndlrLostVillages) cmd() command {
	return cmdLostVillages
}

func (hndlr *hndlrLostVillages) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrLostVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpLostVillages,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := hndlr.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.LostVillagesGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups[0].LostVillagesChannelID = m.ChannelID
	go hndlr.cfg.GroupRepository.Update(context.Background(), groups[0])

	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.LostVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrDisableLostVillages struct {
	*Session
}

var _ commandHandler = &hndlrDisableLostVillages{}

func (hndlr *hndlrDisableLostVillages) cmd() command {
	return cmdDisableLostVillages
}

func (hndlr *hndlrDisableLostVillages) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrDisableLostVillages) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDisableLostVillages,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DisableLostVillagesInvalidID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	groups, _, err := hndlr.cfg.GroupRepository.Fetch(context.Background(), &model.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		hndlr.SendMessage(m.ChannelID,
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
		go hndlr.cfg.GroupRepository.Update(context.Background(), groups[0])
	}

	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.DisableLostVillagesSuccess,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrObserve struct {
	*Session
}

var _ commandHandler = &hndlrObserve{}

func (hndlr *hndlrObserve) cmd() command {
	return cmdObserve
}

func (hndlr *hndlrObserve) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrObserve) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 3 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObserve,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
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
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveInvalidTribeID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	server, err := hndlr.cfg.API.Server.Read(serverKey, nil)
	if err != nil || server == nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveServerNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	if server.Status == twmodel.ServerStatusClosed {
		hndlr.SendMessage(m.ChannelID,
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
		tribe, err = hndlr.cfg.API.Tribe.Read(server.Key, tribeID)
	} else {
		list := &sdk.TribeList{}
		list, err = hndlr.cfg.API.Tribe.Browse(server.Key, 1, 0, []string{}, &twmodel.TribeFilter{
			Tag: []string{tribeTag},
		})
		if list != nil && list.Items != nil && len(list.Items) > 0 {
			tribe = list.Items[0]
		}
	}
	if err != nil || tribe == nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveTribeNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := hndlr.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObserveGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if len(group.Observations) >= observationsPerGroup {
		hndlr.SendMessage(m.ChannelID,
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

	go hndlr.cfg.ObservationRepository.Store(context.Background(), &model.Observation{
		Server:  server.Key,
		TribeID: tribe.ID,
		GroupID: groupID,
	})

	hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.ObserveSuccess,
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

type hndlrDeleteObservation struct {
	*Session
}

var _ commandHandler = &hndlrDeleteObservation{}

func (hndlr *hndlrDeleteObservation) cmd() command {
	return cmdDeleteObservation
}

func (hndlr *hndlrDeleteObservation) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrDeleteObservation) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 2 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpDeleteObservation,
				TemplateData: map[string]interface{}{
					"Command":             hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"ObservationsCommand": cmdObservations.WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand":       cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
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
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationInvalidTribeID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	group, err := hndlr.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.DeleteObservationGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	go hndlr.cfg.ObservationRepository.Delete(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	hndlr.SendMessage(m.ChannelID, ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: message.DeleteObservationSuccess,
		TemplateData: map[string]interface{}{
			"Mention": m.Author.Mention(),
		},
	}))
}

type hndlrObservations struct {
	*Session
}

var _ commandHandler = &hndlrObservations{}

func (hndlr *hndlrObservations) cmd() command {
	return cmdObservations
}

func (hndlr *hndlrObservations) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrObservations) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpObservations,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObservationsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := hndlr.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ObservationsGroupNotFound,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	observations, _, err := hndlr.cfg.ObservationRepository.Fetch(context.Background(), &model.ObservationFilter{
		GroupID: []int{groupID},
		DefaultFilter: model.DefaultFilter{
			Order: []string{"id ASC"},
		},
	})
	if err != nil {
		hndlr.SendMessage(m.ChannelID,
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
		list, err := hndlr.cfg.API.Tribe.Browse(server, 0, 0, []string{}, &twmodel.TribeFilter{
			ID: tribeIDs,
		})
		if err != nil {
			hndlr.SendMessage(m.ChannelID,
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
	versionList, err := hndlr.cfg.API.Version.Browse(0, 0, []string{}, &twmodel.VersionFilter{
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
	hndlr.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ObservationsTitle,
		})).
		SetFields(bldr.ToMessageEmbedFields()))
}

type hndlrShowEnnobledBarbarians struct {
	*Session
}

var _ commandHandler = &hndlrShowEnnobledBarbarians{}

func (hndlr *hndlrShowEnnobledBarbarians) cmd() command {
	return cmdShowEnnobledBarbarians
}

func (hndlr *hndlrShowEnnobledBarbarians) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrShowEnnobledBarbarians) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowEnnobledBarbs,
				TemplateData: map[string]interface{}{
					"Command":       cmdShowEnnobledBarbarians.WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := hndlr.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		hndlr.SendMessage(m.ChannelID,
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
	if err := hndlr.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowEnnobledBarbsSuccess1,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ShowEnnobledBarbsSuccess2,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}

type hndlrShowInternals struct {
	*Session
}

var _ commandHandler = &hndlrShowInternals{}

func (hndlr *hndlrShowInternals) cmd() command {
	return cmdShowInternals
}

func (hndlr *hndlrShowInternals) requireAdmPermissions() bool {
	return true
}

func (hndlr *hndlrShowInternals) execute(ctx *commandCtx, m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength != 1 {
		hndlr.SendMessage(m.ChannelID,
			m.Author.Mention()+" "+ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.HelpShowInternals,
				TemplateData: map[string]interface{}{
					"Command":       hndlr.cmd().WithPrefix(hndlr.cfg.CommandPrefix),
					"GroupsCommand": cmdGroups.WithPrefix(hndlr.cfg.CommandPrefix),
				},
			}))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil || groupID <= 0 {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsInvalidGroupID,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	group, err := hndlr.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		hndlr.SendMessage(m.ChannelID,
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
	if err := hndlr.cfg.GroupRepository.Update(context.Background(), group); err != nil {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.InternalServerError,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}

	if oldValue {
		hndlr.SendMessage(m.ChannelID,
			ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: message.ShowInternalsSuccess1,
				TemplateData: map[string]interface{}{
					"Mention": m.Author.Mention(),
				},
			}))
		return
	}
	hndlr.SendMessage(m.ChannelID,
		ctx.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: message.ShowInternalsSuccess2,
			TemplateData: map[string]interface{}{
				"Mention": m.Author.Mention(),
			},
		}))
}
