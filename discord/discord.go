package discord

import (
	"fmt"
	"strings"

	"github.com/tribalwarshelp/dcbot/observation"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/golang-sdk/sdk"

	"github.com/bwmarrin/discordgo"
)

type SessionConfig struct {
	Token                 string
	CommandPrefix         string
	Status                string
	ServerRepository      server.Repository
	ObservationRepository observation.Repository
	API                   *sdk.SDK
}

type Session struct {
	dg  *discordgo.Session
	cfg SessionConfig
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
	s.dg.AddHandler(s.handleNewMessage)

	err := s.dg.Open()
	if err != nil {
		return fmt.Errorf("error opening ws connection: %s", err.Error())
	}

	if err := s.dg.UpdateStatus(0, s.cfg.Status); err != nil {
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
	_, err := s.dg.ChannelMessageSendEmbed(channelID, message)
	return err
}

func (s *Session) handleNewMessage(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.dg.State.User.ID || m.Author.Bot {
		return
	}

	splitted := strings.Split(m.Content, " ")
	argsLength := len(splitted) - 1
	args := splitted[1 : argsLength+1]
	switch splitted[0] {
	case HelpCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleHelpCommand(m)
	case AuthorCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleAuthorCommand(m)
	case ObserveCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleObserveCommand(m, args...)
	case UnObserveCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleUnObserveCommand(m, args...)
	case ObservationsCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleObservationsCommand(m)
	case ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleConqueredVillagesCommand(m)
	case UnObserveConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleUnObserveConqueredVillagesCommand(m)
	case LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleLostVillagesCommand(m)
	case UnObserveLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleUnObserveLostVillagesCommand(m)
	case TribeCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTribeCommand(m, args...)
	}
}

func (s *Session) memberHasPermission(guildID string, userID string, permission int) (bool, error) {
	member, err := s.dg.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.dg.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	// check if a user is guild owner
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

func (s *Session) sendUnknownCommandError(mention, channelID string, command ...string) {
	s.SendMessage(channelID, mention+` Nieznana komenda: `+strings.Join(command, " "))
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
