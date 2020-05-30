package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"twdcbot/models"
	"twdcbot/server"
	"twdcbot/tribe"

	"github.com/bwmarrin/discordgo"
)

const (
	TribesPerServer = 10
)

type SessionConfig struct {
	Token            string
	CommandPrefix    string
	Status           string
	ServerRepository server.Repository
	TribeRepository  tribe.Repository
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

func (s *Session) handleNewMessage(_ *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.dg.State.User.ID || m.Author.Bot || m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}
	splitted := strings.Split(m.Content, " ")
	argsLength := len(splitted) - 1
	args := splitted[1 : argsLength+1]
	switch splitted[0] {
	case HelpCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleHelpCommand(m)
	case AddCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleAddCommand(m, args...)
	case DeleteCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleDeleteCommand(m, args...)
	case ListCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleListCommand(m)
	case LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
	case ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
	}
}

func (s *Session) handleHelpCommand(m *discordgo.MessageCreate) {
	s.sendHelpMessage(m.Author.Mention(), m.ChannelID)
}

func (s *Session) handleListCommand(m *discordgo.MessageCreate) {
	tribes, _, err := s.cfg.TribeRepository.Fetch(context.Background(), &models.TribeFilter{
		ServerID: []string{m.GuildID},
	})
	if err != nil {
		return
	}
	msg := m.Author.Mention() + " ```ID w bazie - Świat - ID plemienia \n\n"
	for _, tribe := range tribes {
		msg += fmt.Sprintf(">>> %d - %s - %d\n", tribe.ID, tribe.World, tribe.TribeID)
	}
	msg += "```"
	s.sendMessage(m.ChannelID, msg)
}

func (s *Session) handleAddCommand(m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength > 2 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[2:argsLength]...)
		return
	} else if argsLength < 2 {
		s.sendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [świat] [id plemienia]",
				m.Author.Mention(),
				AddCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}
	world := args[0]
	id, err := strconv.Atoi(args[1])
	if err != nil {
		s.sendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [świat] [id plemienia]",
				m.Author.Mention(),
				AddCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}
	server := &models.Server{
		ID: m.GuildID,
	}
	err = s.cfg.ServerRepository.Store(context.Background(), server)
	if err != nil {
		s.sendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}
	if len(server.Tribes) >= TribesPerServer {
		s.sendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` Osiągnięto limit plemion (%d/%d).`, TribesPerServer, TribesPerServer))
		return
	}
	err = s.cfg.TribeRepository.Store(context.Background(), &models.Tribe{
		World:    world,
		TribeID:  id,
		ServerID: server.ID,
	})
	if err != nil {
		s.sendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}

	s.sendMessage(m.ChannelID, m.Author.Mention()+` Dodano.`)
}

func (s *Session) handleDeleteCommand(m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.sendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id z tw!list]`,
				m.Author.Mention(),
				DeleteCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		s.sendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id z tw!list]`,
				m.Author.Mention(),
				DeleteCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	go s.cfg.TribeRepository.Delete(context.Background(), &models.TribeFilter{
		ServerID: []string{m.GuildID},
		ID:       []int{id},
	})

	s.sendMessage(m.ChannelID, m.Author.Mention()+` Usunięto.`)
}

func (s *Session) Close() error {
	return s.dg.Close()
}

func (s *Session) memberHasPermission(guildID string, userID string, permission int) (bool, error) {
	member, err := s.dg.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.dg.GuildMember(guildID, userID); err != nil {
			return false, err
		}
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
