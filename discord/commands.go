package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command string

var (
	HelpCommand              Command = "help"
	AddCommand               Command = "add"
	ListCommand              Command = "list"
	DeleteCommand            Command = "delete"
	LostVillagesCommand      Command = "lostvillages"
	ConqueredVillagesCommand Command = "conqueredvillages"
	TopAttCommand            Command = "topatt"
	TopDefCommand            Command = "topdef"
	TopSuppCommand           Command = "topsupp"
	TopTotalCommand          Command = "toptotal"
	TopPointsCommand         Command = "toppoints"
)

func (cmd Command) String() string {
	return string(cmd)
}

func (cmd Command) WithPrefix(prefix string) string {
	return prefix + cmd.String()
}

func (s *Session) sendHelpMessage(channelID string) {
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       discordEmbedColor,
		Title:       "Pomoc",
		Description: "Komendy oferowane przez bota",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name: "Dla wszystkich",
				Value: fmt.Sprintf(`
- %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RA z plemion o podanych id
- %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RO z plemion o podanych id
- %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RW z plemion o podanych id
- %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie pokonanych z plemion o podanych id
- %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie punktów z plemion o podanych id
				`,
					TopAttCommand.WithPrefix(s.cfg.CommandPrefix),
					TopDefCommand.WithPrefix(s.cfg.CommandPrefix),
					TopSuppCommand.WithPrefix(s.cfg.CommandPrefix),
					TopTotalCommand.WithPrefix(s.cfg.CommandPrefix),
					TopPointsCommand.WithPrefix(s.cfg.CommandPrefix)),
				Inline: false,
			},
			&discordgo.MessageEmbedField{
				Name: "Dla adminów",
				Value: fmt.Sprintf(`
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
					ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix)),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "https://dawid-wysokinski.pl/",
		},
	}
	s.dg.ChannelMessageSendEmbed(channelID, embed)
}

func (s *Session) sendUnknownCommandError(mention, channelID string, command ...string) {
	s.SendMessage(channelID, mention+` Nieznana komenda: `+strings.Join(command, " "))
}

func (s *Session) SendMessage(channelID, message string) {
	s.dg.ChannelMessageSend(channelID, message)
}
