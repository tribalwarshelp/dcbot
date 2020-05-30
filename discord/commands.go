package discord

import (
	"fmt"
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
	s.SendMessage(channelID, mention+"```Dostępne komendy \n"+fmt.Sprintf(`
- %s [świat] [id] - dodaje plemię z danego świata do obserwowanych
- %s - wyświetla wszystkie obserwowane plemiona
- %s [id z %s] - usuwa plemię z obserwowanych
- %s - ustawia kanał na którym będą wyświetlać się informacje o straconych wioskach
- %s - ustawia kanał na którym będą wyświetlać się informacje o podbitych wioskach
	`,
		AddCommand.WithPrefix(s.cfg.CommandPrefix),
		ListCommand.WithPrefix(s.cfg.CommandPrefix),
		DeleteCommand.WithPrefix(s.cfg.CommandPrefix),
		ListCommand.WithPrefix(s.cfg.CommandPrefix),
		LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
		ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix))+"```")
}

func (s *Session) sendUnknownCommandError(mention, channelID string, command ...string) {
	s.SendMessage(channelID, mention+` Nieznana komenda: `+strings.Join(command, " "))
}

func (s *Session) SendMessage(channelID, message string) {
	s.dg.ChannelMessageSend(channelID, message)
}
