package discord

import (
	"strings"
)

type Command string

var (
	HelpCommand              Command = "help"
	AddCommand               Command = "add"
	ListCommand              Command = "list"
	DeleteCommand            Command = "delete"
	LostVillagesCommand      Command = "lostvillages"
	ConqueredVillagesCommand Command = "conqueredvillages"
)

func (cmd Command) String() string {
	return string(cmd)
}

func (cmd Command) WithPrefix(prefix string) string {
	return prefix + cmd.String()
}

func (s *Session) sendHelpMessage(mention, channelID string) {
	s.SendMessage(channelID, mention+" zaraz ogarnę help cmd")
}

func (s *Session) sendUnknownCommandError(mention, channelID string, command ...string) {
	s.SendMessage(channelID, mention+` Nieznana komenda: `+strings.Join(command, " "))
}

func (s *Session) SendMessage(channelID, message string) {
	s.dg.ChannelMessageSend(channelID, message)
}
