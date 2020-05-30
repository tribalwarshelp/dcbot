package discord

import (
	"context"
	"fmt"
	"log"
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
	if has, err := s.MemberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}
	splitted := strings.Split(m.Content, " ")
	argsLength := len(splitted) - 1
	args := splitted[1 : argsLength+1]
	switch splitted[0] {
	case HelpCommand.WithPrefix(s.cfg.CommandPrefix):
		if argsLength == 0 {
			s.sendHelpMessage(m.Author.Mention(), m.ChannelID)
		} else {
			s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args...)
		}
		break
	case AddCommand.WithPrefix(s.cfg.CommandPrefix):
		if argsLength > 2 {
			s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, splitted[3:argsLength+1]...)
			return
		} else if argsLength < 2 {
			s.sendMessage(m.ChannelID, m.Author.Mention()+` tw!add [świat] [id_plemienia]`)
			return
		}
		s.handleAddCommand(m, args...)
	}
}

func (s *Session) handleAddCommand(m *discordgo.MessageCreate, args ...string) {
	world := args[0]
	id, err := strconv.Atoi(args[1])
	if err != nil {
		s.sendMessage(m.ChannelID, m.Author.Mention()+` tw!add [świat] [id_plemienia]`)
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
	log.Println(world, id, server.Tribes)
	err = s.cfg.TribeRepository.Store(context.Background(), &models.Tribe{
		World:    world,
		TribeID:  id,
		ServerID: server.ID,
	})
	if err != nil {
		s.sendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}
}

func (s *Session) Close() error {
	return s.dg.Close()
}

func (s *Session) MemberHasPermission(guildID string, userID string, permission int) (bool, error) {
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
