package discord

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/server"
	"github.com/tribalwarshelp/dcbot/tribe"
	"github.com/tribalwarshelp/golang-sdk/sdk"
	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/bwmarrin/discordgo"
)

const (
	TribesPerServer   = 10
	discordEmbedColor = 0x00ff00
)

type SessionConfig struct {
	Token            string
	CommandPrefix    string
	Status           string
	ServerRepository server.Repository
	TribeRepository  tribe.Repository
	API              *sdk.SDK
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
	if m.Author.ID == s.dg.State.User.ID || m.Author.Bot {
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
		s.handleLostVillagesCommand(m)
	case ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleConqueredVillagesCommand(m)
	case TopAttCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTopCommands(m, TopAttCommand, args...)
	case TopDefCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTopCommands(m, TopDefCommand, args...)
	case TopSuppCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTopCommands(m, TopSuppCommand, args...)
	case TopTotalCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTopCommands(m, TopTotalCommand, args...)
	case TopPointsCommand.WithPrefix(s.cfg.CommandPrefix):
		s.handleTopCommands(m, TopPointsCommand, args...)
	}
}

func (s *Session) handleHelpCommand(m *discordgo.MessageCreate) {
	s.sendHelpMessage(m.ChannelID)
}

func (s *Session) handleTopCommands(m *discordgo.MessageCreate, command Command, args ...string) {
	argsLength := len(args)
	if argsLength < 3 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [świat] [strona] [id...]",
				m.Author.Mention(),
				command.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	world := args[0]
	page, err := strconv.Atoi(args[1])
	if err != nil || page <= 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s 2 argument musi być liczbą większą od 0.", m.Author.Mention()))
		return
	}
	ids := []int{}
	for _, arg := range args[2:argsLength] {
		id, err := strconv.Atoi(arg)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie wprowadziłeś ID plemion.", m.Author.Mention()))
		return
	}

	exist := true
	limit := 10
	offset := (page - 1) * limit
	filter := &shared_models.PlayerFilter{
		Exist:   &exist,
		TribeID: ids,
		Limit:   limit,
		Offset:  offset,
	}
	title := ""
	switch command {
	case TopAttCommand:
		filter.RankAttGTE = 1
		filter.Sort = "rankAtt ASC"
		title = "Top pokonani w ataku"
	case TopDefCommand:
		filter.RankDefGTE = 1
		filter.Sort = "rankDef ASC"
		title = "Top pokonani w obronie"
	case TopSuppCommand:
		filter.RankSupGTE = 1
		filter.Sort = "rankSup ASC"
		title = "Top pokonani jako wspierający"
	case TopTotalCommand:
		filter.RankTotalGTE = 1
		filter.Sort = "rankTotal ASC"
		title = "Top pokonani ogólnie"
	case TopPointsCommand:
		filter.Sort = "rank ASC"
		title = "Najwięcej punktów"
	}

	playersList, err := s.cfg.API.Players.Browse(world, filter, &sdk.PlayerInclude{
		Tribe: true,
	})
	if err != nil {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie udało się wygenerować listy.", m.Author.Mention()))
		return
	}
	if playersList.Total == 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie znaleziono plemion o podanych ID.", m.Author.Mention()))
		return
	}

	msg := ""
	for i, player := range playersList.Items {
		rank := 0
		score := 0
		switch command {
		case TopAttCommand:
			rank = player.RankAtt
			score = player.ScoreAtt
		case TopDefCommand:
			rank = player.RankDef
			score = player.ScoreDef
		case TopSuppCommand:
			rank = player.RankSup
			score = player.ScoreSup
		case TopTotalCommand:
			rank = player.RankTotal
			score = player.ScoreTotal
		case TopPointsCommand:
			rank = player.Rank
			score = player.Points
		}

		msg += fmt.Sprintf("**%d**. **__%s__** (Plemię: **%s** | Ranking ogólny: **%d** | Wynik: **%d**)\n",
			offset+i+1,
			player.Name,
			player.Tribe.Tag,
			rank, score)
	}

	totalPages := int(math.Round(float64(playersList.Total) / float64(limit)))
	s.dg.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       discordEmbedColor,
		Title:       title,
		Description: "A oto lista!",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "-",
				Value:  msg,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Strona %d z %d", page, totalPages),
		},
	})
}

func (s *Session) handleLostVillagesCommand(m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	server := &models.Server{
		ID: m.GuildID,
	}
	err := s.cfg.ServerRepository.Store(context.Background(), server)
	if err != nil {
		return
	}
	server.LostVillagesChannelID = m.ChannelID
	go s.cfg.ServerRepository.Update(context.Background(), server)
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Pomyślnie zmieniono kanał na którym będą się wyświetlać informacje o straconych wioskach.", m.Author.Mention()))
}

func (s *Session) handleConqueredVillagesCommand(m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	server := &models.Server{
		ID: m.GuildID,
	}
	err := s.cfg.ServerRepository.Store(context.Background(), server)
	if err != nil {
		return
	}
	server.ConqueredVillagesChannelID = m.ChannelID
	go s.cfg.ServerRepository.Update(context.Background(), server)
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Pomyślnie zmieniono kanał na którym będą się wyświetlać informacje o podbitych wioskach.", m.Author.Mention()))
}

func (s *Session) handleAddCommand(m *discordgo.MessageCreate, args ...string) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 2 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[2:argsLength]...)
		return
	} else if argsLength < 2 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [świat] [id plemienia]",
				m.Author.Mention(),
				AddCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	world := args[0]
	id, err := strconv.Atoi(args[1])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [świat] [id plemienia]",
				m.Author.Mention(),
				AddCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	tribe, err := s.cfg.API.Tribes.Read(world, id)
	if err != nil || tribe == nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` Plemię o ID: %d nie istnieje na świecie %s.`, id, world))
		return
	}

	server := &models.Server{
		ID: m.GuildID,
	}
	err = s.cfg.ServerRepository.Store(context.Background(), server)
	if err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}

	if len(server.Tribes) >= TribesPerServer {
		s.SendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` Osiągnięto limit plemion (%d/%d).`, TribesPerServer, TribesPerServer))
		return
	}

	err = s.cfg.TribeRepository.Store(context.Background(), &models.Tribe{
		World:    world,
		TribeID:  id,
		ServerID: server.ID,
	})
	if err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}

	s.SendMessage(m.ChannelID, m.Author.Mention()+` Dodano.`)
}

func (s *Session) handleDeleteCommand(m *discordgo.MessageCreate, args ...string) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 1 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[1:argsLength]...)
		return
	} else if argsLength < 1 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id z tw!list]`,
				m.Author.Mention(),
				DeleteCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s %s [id z tw!list]`,
				m.Author.Mention(),
				DeleteCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	go s.cfg.TribeRepository.Delete(context.Background(), &models.TribeFilter{
		ServerID: []string{m.GuildID},
		ID:       []int{id},
	})

	s.SendMessage(m.ChannelID, m.Author.Mention()+` Usunięto.`)
}

func (s *Session) handleListCommand(m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	tribes, _, err := s.cfg.TribeRepository.Fetch(context.Background(), &models.TribeFilter{
		ServerID: []string{m.GuildID},
	})
	if err != nil {
		return
	}

	msg := ""
	for i, tribe := range tribes {
		msg += fmt.Sprintf("**%d**. %d - %s - %d\n", i+1, tribe.ID, tribe.World, tribe.TribeID)
	}

	s.dg.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Title:  "Lista obserwowanych plemion",
		Color:  discordEmbedColor,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  "Indeks. ID - świat - ID plemienia",
				Value: msg,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Strona 1 z 1",
		},
	})
}

func (s *Session) Close() error {
	return s.dg.Close()
}

func (s *Session) memberHasPermission(guildID string, userID string, permission int) (bool, error) {
	guild, err := s.dg.State.Guild(guildID)
	if err != nil {
		if guild, err = s.dg.Guild(guildID); err != nil {
			return false, err
		}
	}
	if guild.OwnerID == userID {
		return true, nil
	}

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
