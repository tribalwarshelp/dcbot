package discord

import (
	"context"
	"fmt"
	"math"
	"strconv"

	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/bwmarrin/discordgo"
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

const (
	TribesPerServer = 10
)

type Command string

const (
	HelpCommand              Command = "help"
	AddCommand               Command = "add"
	ListCommand              Command = "list"
	DeleteCommand            Command = "delete"
	LostVillagesCommand      Command = "lostvillages"
	ConqueredVillagesCommand Command = "conqueredvillages"
	TribeCommand             Command = "tribe"
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

func (s *Session) handleHelpCommand(m *discordgo.MessageCreate) {
	tribeCMDWithPrefix := TribeCommand.WithPrefix(s.cfg.CommandPrefix)
	commandsForAll := fmt.Sprintf(`
- %s %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RA z plemion o podanych id
- %s %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RO z plemion o podanych id
- %s %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RW z plemion o podanych id
- %s %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie pokonanych z plemion o podanych id
- %s %s [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie punktów z plemion o podanych id
				`,
		tribeCMDWithPrefix,
		TopAttCommand.String(),
		tribeCMDWithPrefix,
		TopDefCommand.String(),
		tribeCMDWithPrefix,
		TopSuppCommand.String(),
		tribeCMDWithPrefix,
		TopTotalCommand.String(),
		tribeCMDWithPrefix,
		TopPointsCommand.String())

	commandsForGuildAdmins := fmt.Sprintf(`
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
		ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix))

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Pomoc").
		SetDescription("Komendy oferowane przez bota").
		AddField("Dla wszystkich", commandsForAll).
		AddField("Dla adminów", commandsForGuildAdmins).
		MessageEmbed)
}

func (s *Session) handleTribeCommand(m *discordgo.MessageCreate, args ...string) {
	argsLength := len(args)
	if argsLength < 4 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawna komenda (sprawdź %s)",
				m.Author.Mention(),
				HelpCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	command := Command(args[0])
	world := args[1]
	page, err := strconv.Atoi(args[2])
	if err != nil || page <= 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s 3 argument musi być liczbą większą od 0.", m.Author.Mention()))
		return
	}
	ids := []int{}
	for _, arg := range args[3:argsLength] {
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
	default:
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Nieznana komenda %s (sprawdź %s)",
				m.Author.Mention(),
				command.String(),
				HelpCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
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

	langVersion, err := s.cfg.API.LangVersions.Read(utils.LanguageCodeFromWorldName(world))
	if err != nil {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie udało się wygenerować listy.", m.Author.Mention()))
		return
	}

	msg := &EmbedMessage{}
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

		msg.Append(fmt.Sprintf("**%d**. [%s](%s) (Plemię: [%s](%s) | Ranking ogólny: **%d** | Wynik: **%d**)\n",
			offset+i+1,
			player.Name,
			utils.FormatPlayerURL(world, langVersion.Host, player.ID),
			player.Tribe.Tag,
			utils.FormatTribeURL(world, langVersion.Host, player.Tribe.ID),
			rank,
			score))
	}

	totalPages := int(math.Round(float64(playersList.Total) / float64(limit)))
	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(title).
		SetDescription("A oto lista!").
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter(fmt.Sprintf("Strona %d z %d", page, totalPages)).
		MessageEmbed)
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

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Lista obserwowanych plemion").
		AddField("Indeks. ID - świat - ID plemienia", msg).
		SetFooter("Strona 1 z 1").
		MessageEmbed)
}
