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

var log = logrus.WithField("package", "discord")

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
	dg       *discordgo.Session
	cfg      SessionConfig
	handlers commandHandlers
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
		&commandHandler{
			cmd: HelpCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleHelpCommand,
		},
		&commandHandler{
			cmd: AuthorCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleAuthorCommand,
		},
		&commandHandler{
			cmd: TribeCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleTribeCommand,
		},
		&commandHandler{
			cmd:                   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleChangeLanguageCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   AddGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleAddGroupCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDeleteGroupCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleGroupsCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleObserveCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   DeleteObservationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDeleteObservationCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleObservationsCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleConqueredVillagesCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   DisableConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableConqueredVillagesCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleLostVillagesCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   DisableLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableLostVillagesCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleShowEnnobledBarbariansCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   ShowInternalsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleShowInternalsCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   CoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleCoordsTranslationCommand,
			requireAdmPermissions: true,
		},
		&commandHandler{
			cmd:                   DisableCoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableCoordsTranslationCommand,
			requireAdmPermissions: true,
		},
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

func (s *Session) SendEmbed(channelID string, message *discordgo.MessageEmbed) error {
	fields := message.Fields
	baseNumberOfCharacters := len(message.Description) + len(message.Title)
	if message.Author != nil {
		baseNumberOfCharacters += len(message.Author.Name)
	}
	if message.Footer != nil {
		baseNumberOfCharacters += len(message.Footer.Text)
	}

	var splittedFields [][]*discordgo.MessageEmbedField
	characters := baseNumberOfCharacters
	fromIndex := 0
	fieldsLen := len(fields)
	for index, field := range fields {
		fNameLen := len(field.Name)
		fValLen := len(field.Value)
		if characters+fNameLen+fValLen > EmbedSizeLimit || index == fieldsLen-1 {
			splittedFields = append(splittedFields, fields[fromIndex:index+1])
			fromIndex = index
			characters = baseNumberOfCharacters
		}
		characters += fNameLen
		characters += fValLen
	}
	for _, fields := range splittedFields {
		fieldsLen := len(fields)
		for i := 0; i < fieldsLen; i += EmbedLimitField {
			end := i + EmbedLimitField

			if end > fieldsLen {
				end = fieldsLen
			}
			message.Fields = fields[i:end]
			if _, err := s.dg.ChannelMessageSendEmbed(channelID, message); err != nil {
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

	cmd := Command(parts[0])
	h := s.handlers.find(cmd)
	if h != nil {
		if h.requireAdmPermissions {
			if m.GuildID == "" {
				return
			}
			has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator)
			if err != nil || !has {
				return
			}
		}
		log.
			WithFields(logrus.Fields{
				"serverID":       svr.ID,
				"lang":           svr.Lang,
				"command":        cmd,
				"args":           args,
				"authorID":       m.Author.ID,
				"authorUsername": m.Author.Username,
			}).
			Info("handleNewMessage: Executing command...")
		h.fn(ctx, m, args...)
		return
	}

	s.translateCoords(ctx, m)
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
