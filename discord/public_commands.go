package discord

import (
	"fmt"
	"math"
	"strconv"

	shared_models "github.com/tribalwarshelp/shared/models"

	"github.com/bwmarrin/discordgo"
	"github.com/tribalwarshelp/dcbot/utils"
	"github.com/tribalwarshelp/golang-sdk/sdk"
)

const (
	HelpCommand      Command = "help"
	TribeCommand     Command = "tribe"
	TopAttCommand    Command = "topatt"
	TopDefCommand    Command = "topdef"
	TopSuppCommand   Command = "topsupp"
	TopTotalCommand  Command = "toptotal"
	TopPointsCommand Command = "toppoints"
	AuthorCommand    Command = "author"
)

func (s *Session) handleHelpCommand(m *discordgo.MessageCreate) {
	tribeCMDWithPrefix := TribeCommand.WithPrefix(s.cfg.CommandPrefix)
	commandsForAll := fmt.Sprintf(`
- **%s %s** [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RA z plemion o podanych id
- **%s %s** [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RO z plemion o podanych id
- **%s %s** [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największym RW z plemion o podanych id
- **%s %s** [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie pokonanych z plemion o podanych id
- **%s %s** [serwer] [strona] [id1] [id2] [id3] [n id] - wyświetla graczy o największej liczbie punktów z plemion o podanych id
- **%s** - kontakt z autorem bota
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
		TopPointsCommand.String(),
		AuthorCommand.WithPrefix(s.cfg.CommandPrefix),
	)

	commandsForGuildAdmins := fmt.Sprintf(`
- **%s** - tworzy nową grupę
- **%s** - lista grup
- **%s** [id grupy z %s] - usuwa grupę
- **%s** [id grupy z %s] - włącza/wyłącza wyświetlanie powiadomień o podbitych wioskach barbarzyńskich
- **%s** [id grupy z %s] [świat] [id plemienia] - dodaje plemię z danego świata do obserwowanych
- **%s** [id grupy z %s] - wyświetla wszystkie obserwowane plemiona
- **%s** [id grupy z %s] [id z %s] - usuwa plemię z obserwowanych
- **%s** [id grupy z %s] - ustawia kanał na którym będą wyświetlać się informacje o podbitych wioskach
- **%s** [id grupy z %s] - informacje o podbitych wioskach na wybranym kanale nie będą się już pojawiały
- **%s** [id grupy z %s] - ustawia kanał na którym będą wyświetlać się informacje o straconych wioskach
- **%s** [id grupy z %s] - informacje o straconych wioskach na wybranym kanale nie będą się już pojawiały
				`,
		AddGroupCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		ShowEnnobledBarbariansCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		ObserveCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		UnObserveCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		ObservationsCommand.WithPrefix(s.cfg.CommandPrefix),
		ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		UnObserveConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
		UnObserveLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix),
		GroupsCommand.WithPrefix(s.cfg.CommandPrefix),
	)

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Pomoc").
		SetDescription("Komendy oferowane przez bota").
		AddField("Dla wszystkich", commandsForAll).
		AddField("Dla adminów", commandsForGuildAdmins).
		MessageEmbed)
}

func (s *Session) handleAuthorCommand(m *discordgo.MessageCreate) {
	s.SendMessage(m.ChannelID, fmt.Sprintf("%s Discord: Kichiyaki#2064 | https://dawid-wysokinski.pl/#contact.", m.Author.Mention()))
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
		if err != nil || id <= 0 {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie wprowadziłeś ID plemion.", m.Author.Mention()))
		return
	}

	exists := true
	limit := 10
	offset := (page - 1) * limit
	filter := &shared_models.PlayerFilter{
		Exists:  &exists,
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
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Wystąpił błąd podczas pobierania danych z API, prosimy spróbować później.", m.Author.Mention()))
		return
	}
	if playersList == nil || playersList.Total == 0 {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie znaleziono graczy należących do plemion o podanych ID.", m.Author.Mention()))
		return
	}
	totalPages := int(math.Ceil(float64(playersList.Total) / float64(limit)))
	if page > totalPages {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Przekroczyłeś limit stron (%d/%d).", m.Author.Mention(), page, totalPages))
		return
	}

	langTag := utils.LanguageTagFromWorldName(world)
	langVersion, err := s.cfg.API.LangVersions.Read(langTag)
	if err != nil || langVersion == nil {
		s.SendMessage(m.ChannelID, fmt.Sprintf("%s Nie znaleziono wersji językowej: %s.", m.Author.Mention(), langTag))
		return
	}

	msg := &EmbedMessage{}
	for i, player := range playersList.Items {
		if player == nil {
			continue
		}

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

		tribeTag := "-"
		tribeURL := "-"
		if player.Tribe != nil {
			tribeTag = player.Tribe.Tag
			tribeURL = utils.FormatTribeURL(world, langVersion.Host, player.Tribe.ID)
		}

		msg.Append(fmt.Sprintf("**%d**. [``%s``](%s) (Plemię: [``%s``](%s) | Ranking ogólny: **%d** | Wynik: **%d**)\n",
			offset+i+1,
			player.Name,
			utils.FormatPlayerURL(world, langVersion.Host, player.ID),
			tribeTag,
			tribeURL,
			rank,
			score))
	}

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle(title).
		SetDescription("A oto lista!").
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter(fmt.Sprintf("Strona %d z %d", page, totalPages)).
		MessageEmbed)
}
