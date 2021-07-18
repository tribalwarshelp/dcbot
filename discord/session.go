package discord

import (
	"context"
	"github.com/pkg/errors"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/model"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"

	"github.com/bwmarrin/discordgo"
)

type SessionConfig struct {
	Token                 string
	CommandPrefix         string
	Status                string
	ServerRepository      server.Repository
	GroupRepository       group.Repository
	ObservationRepository observation.Repository
	API                   *sdk.SDK
}

type Session struct {
	dg                *discordgo.Session
	cfg               SessionConfig
	handlers          commandHandlers
	messageProcessors []messageProcessor
}

func New(cfg SessionConfig) (*Session, error) {
	var err error
	s := &Session{
		cfg: cfg,
	}
	s.dg, err = discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, err
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) init() error {
	s.handlers = commandHandlers{
		&hndlrHelp{s},
		&hndlrAuthor{s},
		&hndlrTribe{s},
		&hndlrChangeLanguage{s},
		&hndlrAddGroup{s},
		&hndlrDeleteGroup{s},
		&hndlrGroups{s},
		&hndlrObserve{s},
		&hndlrDeleteObservation{s},
		&hndlrObservations{s},
		&hndlrConqueredVillages{s},
		&hndlrDisableConqueredVillages{s},
		&hndlrLostVillages{s},
		&hndlrDisableLostVillages{s},
		&hndlrShowEnnobledBarbarians{s},
		&hndlrShowInternals{s},
		&hndlrCoordsTranslation{s},
		&hndlrDisableCoordsTranslation{s},
	}
	s.messageProcessors = []messageProcessor{
		&procTranslateCoords{s},
	}

	s.dg.AddHandler(s.handleNewMessage)

	err := s.dg.Open()
	if err != nil {
		return errors.Wrap(err, "error opening ws connection")
	}

	if err := s.UpdateStatus(s.cfg.Status); err != nil {
		return err
	}

	return nil
}

func (s *Session) Close() error {
	return s.dg.Close()
}

func (s *Session) SendMessage(channelID, message string) error {
	_, err := s.dg.ChannelMessageSend(channelID, message)
	return err
}

func (s *Session) SendEmbed(channelID string, e *Embed) error {
	for _, fields := range splitEmbedFields(e) {
		fieldsLen := len(fields)
		for i := 0; i < fieldsLen; i += EmbedLimitField {
			end := i + EmbedLimitField
			if end > fieldsLen {
				end = fieldsLen
			}
			e.Fields = fields[i:end]
			if _, err := s.dg.ChannelMessageSendEmbed(channelID, e.MessageEmbed); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Session) UpdateStatus(status string) error {
	if err := s.dg.UpdateStatus(0, status); err != nil {
		return err
	}
	return nil
}

func (s *Session) IsGuildMember(guildID string) (bool, error) {
	_, err := s.dg.State.Guild(guildID)
	if err != nil {
		if _, err = s.dg.Guild(guildID); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *Session) handleNewMessage(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.dg.State.User.ID || m.Author.Bot {
		return
	}

	parts := strings.Split(m.Content, " ")
	args := parts[1:]
	svr := &model.Server{
		ID:   m.GuildID,
		Lang: message.GetDefaultLanguage().String(),
	}
	if svr.ID != "" {
		if err := s.cfg.ServerRepository.Store(context.Background(), svr); err != nil {
			return
		}
	}
	ctx := &commandCtx{
		server:    svr,
		localizer: message.NewLocalizer(svr.Lang),
	}

	h := s.handlers.find(s.cfg.CommandPrefix, parts[0])
	if h != nil {
		if h.requireAdmPermissions() && m.GuildID != "" {
			has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator)
			if err != nil || !has {
				return
			}
		}
		log.
			WithFields(logrus.Fields{
				"serverID":       svr.ID,
				"lang":           svr.Lang,
				"command":        parts[0],
				"args":           args,
				"authorID":       m.Author.ID,
				"authorUsername": m.Author.Username,
			}).
			Infof(`handleNewMessage: Executing command "%s"...`, m.Content)
		h.execute(ctx, m, args...)
		return
	}

	for _, p := range s.messageProcessors {
		p.process(ctx, m)
	}
}

func (s *Session) memberHasPermission(guildID string, userID string, permission int) (bool, error) {
	member, err := s.dg.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.dg.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	// check if user is a guild owner
	guild, err := s.dg.State.Guild(guildID)
	if err != nil {
		if guild, err = s.dg.Guild(guildID); err != nil {
			return false, err
		}
	}
	if guild.OwnerID == userID {
		return true, nil
	}

	// Iterate through the role IDs stored in member.Roles
	// to check permissions
	for _, roleID := range member.Roles {
		role, err := s.dg.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}
	}

	return false, nil
}
