package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tribalwarshelp/dcbot/message"

	"github.com/tribalwarshelp/dcbot/group"
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/bwmarrin/discordgo"
)

type handler struct {
	cmd                   Command
	requireAdmPermissions bool
	fn                    func(ctx *commandCtx, m *discordgo.MessageCreate, args ...string)
}

type handlers []*handler

func (hs handlers) find(cmd Command) *handler {
	for _, h := range hs {
		if h.cmd == cmd {
			return h
		}
	}
	return nil
}

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
	handlers handlers
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
	s.handlers = handlers{
		&handler{
			cmd: HelpCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleHelpCommand,
		},
		&handler{
			cmd: AuthorCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleAuthorCommand,
		},
		&handler{
			cmd: TribeCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:  s.handleTribeCommand,
		},
		&handler{
			cmd:                   ChangeLanguageCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleChangeLanguageCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   AddGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleAddGroupCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDeleteGroupCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleGroupsCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleObserveCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   DeleteObservationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDeleteObservationCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleObservationsCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleConqueredVillagesCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   DisableConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableConqueredVillagesCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleLostVillagesCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   DisableLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableLostVillagesCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleShowEnnobledBarbariansCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   ShowInternalsCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleShowInternalsCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   CoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleCoordsTranslationCommand,
			requireAdmPermissions: true,
		},
		&handler{
			cmd:                   DisableCoordsTranslationCommand.WithPrefix(s.cfg.CommandPrefix),
			fn:                    s.handleDisableCoordsTranslationCommand,
			requireAdmPermissions: true,
		},
	}

	s.dg.AddHandler(s.handleNewMessage)

	err := s.dg.Open()
	if err != nil {
		return fmt.Errorf("error opening ws connection: %s", err.Error())
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

	splittedFields := [][]*discordgo.MessageEmbedField{}
	characters := baseNumberOfCharacters
	fromIndex := 0
	for index, field := range fields {
		if characters+len(field.Name)+len(field.Value) > EmbedLimit || index == len(fields)-1 {
			splittedFields = append(splittedFields, fields[fromIndex:index+1])
			fromIndex = index
			characters = baseNumberOfCharacters
		}
		characters += len(field.Name)
		characters += len(field.Value)
	}
	for _, fields := range splittedFields {
		for i := 0; i < len(fields); i += EmbedLimitField {
			end := i + EmbedLimitField

			if end > len(fields) {
				end = len(fields)
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

	splitted := strings.Split(m.Content, " ")
	argsLength := len(splitted) - 1
	args := splitted[1 : argsLength+1]
	server := &models.Server{
		ID:   m.GuildID,
		Lang: message.GetDefaultLanguage().String(),
	}
	if server.ID != "" {
		if err := s.cfg.ServerRepository.Store(context.Background(), server); err != nil {
			return
		}
	}
	ctx := &commandCtx{
		server:    server,
		localizer: message.NewLocalizer(server.Lang),
	}

	h := s.handlers.find(Command(splitted[0]))
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
		log.Print(h.cmd)
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
