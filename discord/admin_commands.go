package discord

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/tribalwarshelp/dcbot/models"
	"github.com/tribalwarshelp/dcbot/utils"
	shared_models "github.com/tribalwarshelp/shared/models"
)

const (
	ObservationsPerGroup = 10
	GroupsPerServer      = 5
)

const (
	AddGroupCommand                   Command = "addgroup"
	DeleteGroupCommand                Command = "deletegroup"
	GroupsCommand                     Command = "groups"
	ObserveCommand                    Command = "observe"
	ObservationsCommand               Command = "observations"
	UnObserveCommand                  Command = "unobserve"
	LostVillagesCommand               Command = "lostvillages"
	UnObserveLostVillagesCommand      Command = "unobservelostvillages"
	ConqueredVillagesCommand          Command = "conqueredvillages"
	UnObserveConqueredVillagesCommand Command = "unobserveconqueredvillages"
)

func (s *Session) handleAddGroupCommand(m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}

	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	server := &models.Server{
		ID: m.GuildID,
	}
	if err := s.cfg.ServerRepository.Store(context.Background(), server); err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Nie udało się dodać grupy", m.Author.Mention()))
		return
	}
	if len(server.Groups) >= GroupsPerServer {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+fmt.Sprintf(` Osiągnięto limit grup na serwerze (%d/%d).`, GroupsPerServer, GroupsPerServer))
		return
	}

	group := &models.Group{
		ServerID: m.GuildID,
	}
	if err := s.cfg.GroupRepository.Store(context.Background(), group); err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Nie udało się dodać grupy", m.Author.Mention()))
		return
	}

	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Utworzono nową grupę o ID %d.", m.Author.Mention(), group.ID))
}
func (s *Session) handleDeleteGroupCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf("%s %s [id grupy]",
				m.Author.Mention(),
				DeleteGroupCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawne ID grupy (powinna to być liczba całkowita większa od 1).", m.Author.Mention()))
		return
	}

	go s.cfg.GroupRepository.Delete(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})

	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Usunięto grupę.", m.Author.Mention()))
}

func (s *Session) handleGroupsCommand(m *discordgo.MessageCreate) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ServerID: []string{m.GuildID},
	})
	if err != nil {
		return
	}

	msg := ""
	for i, groups := range groups {
		msg += fmt.Sprintf("**%d** | %d\n", i+1, groups.ID)
	}

	if msg == "" {
		msg = "Brak dodanych grup"
	}

	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Lista grup").
		AddField("Indeks | ID", msg).
		SetFooter("Strona 1 z 1").
		MessageEmbed)
}

func (s *Session) handleConqueredVillagesCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf("%s %s [id grupy]",
				m.Author.Mention(),
				ConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawne ID grupy (powinna to być liczba całkowita większa od 1).", m.Author.Mention()))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s nie znaleziono grupy.", m.Author.Mention()))
		return
	}

	groups[0].ConqueredVillagesChannelID = m.ChannelID
	go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Pomyślnie zmieniono kanał na którym będą się wyświetlać informacje o podbitych wioskach (Grupa: %d).",
			m.Author.Mention(), groupID))
}

func (s *Session) handleUnObserveConqueredVillagesCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf("%s %s [id grupy]",
				m.Author.Mention(),
				UnObserveConqueredVillagesCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawne ID grupy (powinna to być liczba całkowita większa od 1).", m.Author.Mention()))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s nie znaleziono grupy.", m.Author.Mention()))
		return
	}

	if groups[0].ConqueredVillagesChannelID != "" {
		groups[0].ConqueredVillagesChannelID = ""
		go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	}
	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Informacje o podbitych wioskach grupy %d nie będą się już pojawiały.", m.Author.Mention(), groupID))
}

func (s *Session) handleLostVillagesCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf("%s %s [id grupy]",
				m.Author.Mention(),
				LostVillagesCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawne ID grupy", m.Author.Mention()))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Nie znaleziono grupy.", m.Author.Mention()))
		return
	}
	groups[0].LostVillagesChannelID = m.ChannelID
	go s.cfg.GroupRepository.Update(context.Background(), groups[0])

	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Pomyślnie zmieniono kanał na którym będą się wyświetlać informacje o straconych wioskach (Grupa: %d).",
			m.Author.Mention(),
			groupID))
}

func (s *Session) handleUnObserveLostVillagesCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf("%s %s [id grupy]",
				m.Author.Mention(),
				UnObserveLostVillagesCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Niepoprawne ID grupy", m.Author.Mention()))
		return
	}

	groups, _, err := s.cfg.GroupRepository.Fetch(context.Background(), &models.GroupFilter{
		ID:       []int{groupID},
		ServerID: []string{m.GuildID},
	})
	if err != nil || len(groups) == 0 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s Nie znaleziono grupy.", m.Author.Mention()))
		return
	}

	if groups[0].LostVillagesChannelID != "" {
		groups[0].LostVillagesChannelID = ""
		go s.cfg.GroupRepository.Update(context.Background(), groups[0])
	}

	s.SendMessage(m.ChannelID,
		fmt.Sprintf("%s Informacje o straconych wioskach grupy %d nie będą się już pojawiały.",
			m.Author.Mention(),
			groupID))
}

func (s *Session) handleObserveCommand(m *discordgo.MessageCreate, args ...string) {
	if m.GuildID == "" {
		return
	}
	if has, err := s.memberHasPermission(m.GuildID, m.Author.ID, discordgo.PermissionAdministrator); err != nil || !has {
		return
	}

	argsLength := len(args)
	if argsLength > 3 {
		s.sendUnknownCommandError(m.Author.Mention(), m.ChannelID, args[3:argsLength]...)
		return
	} else if argsLength < 3 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s %s [id grupy] [świat] [id plemienia]",
				m.Author.Mention(),
				ObserveCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s ID grupy powinno być liczbą całkowitą większą od 0.",
				m.Author.Mention()))
		return
	}
	serverKey := args[1]
	tribeID, err := strconv.Atoi(args[2])
	if err != nil || tribeID <= 0 {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf("%s ID plemienia powinno być liczbą całkowitą większą od 0.",
				m.Author.Mention()))
		return
	}

	server, err := s.cfg.API.Servers.Read(serverKey, nil)
	if err != nil || server == nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` świat %s jest nieobsługiwany.`, serverKey))
		return
	}
	if server.Status == shared_models.ServerStatusClosed {
		s.SendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` świat %s jest zamknięty.`, serverKey))
		return
	}

	tribe, err := s.cfg.API.Tribes.Read(server.Key, tribeID)
	if err != nil || tribe == nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+fmt.Sprintf(` Plemię o ID: %d nie istnieje na świecie %s.`, tribeID, server.Key))
		return
	}

	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}

	if len(group.Observations) >= ObservationsPerGroup {
		s.SendMessage(m.ChannelID,
			m.Author.Mention()+fmt.Sprintf(` Osiągnięto limit plemion w grupie (%d/%d).`, ObservationsPerGroup, ObservationsPerGroup))
		return
	}

	err = s.cfg.ObservationRepository.Store(context.Background(), &models.Observation{
		Server:  server.Key,
		TribeID: tribeID,
		GroupID: groupID,
	})
	if err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie udało się dodać plemienia do obserwowanych.`)
		return
	}

	s.SendMessage(m.ChannelID, m.Author.Mention()+` Dodano.`)
}

func (s *Session) handleUnObserveCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf(`%s %s [id grupy] [id obserwacji]`,
				m.Author.Mention(),
				UnObserveCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID grupy powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}
	observationID, err := strconv.Atoi(args[1])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID obserwacji powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}

	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}

	go s.cfg.ObservationRepository.Delete(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
		ID:      []int{observationID},
	})

	s.SendMessage(m.ChannelID, m.Author.Mention()+` Usunięto.`)
}

func (s *Session) handleObservationsCommand(m *discordgo.MessageCreate, args ...string) {
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
			fmt.Sprintf(`%s %s [id grupy]`,
				m.Author.Mention(),
				ObservationsCommand.WithPrefix(s.cfg.CommandPrefix)))
		return
	}

	groupID, err := strconv.Atoi(args[0])
	if err != nil {
		s.SendMessage(m.ChannelID,
			fmt.Sprintf(`%s ID grupy powinno być liczbą całkowitą większą od 0.`,
				m.Author.Mention()))
		return
	}
	group, err := s.cfg.GroupRepository.GetByID(context.Background(), groupID)
	if err != nil || group.ServerID != m.GuildID {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Nie znaleziono grupy.`)
		return
	}
	observations, _, err := s.cfg.ObservationRepository.Fetch(context.Background(), &models.ObservationFilter{
		GroupID: []int{groupID},
	})
	if err != nil {
		s.SendMessage(m.ChannelID, m.Author.Mention()+` Wystąpił błąd wewnętrzny, prosimy spróbować później.`)
		return
	}

	tribeIDsByServer := make(map[string][]int)
	observationIndexByTribeID := make(map[int]int)
	langTags := []shared_models.LanguageTag{}
	for i, observation := range observations {
		tribeIDsByServer[observation.Server] = append(tribeIDsByServer[observation.Server], observation.TribeID)
		observationIndexByTribeID[observation.TribeID] = i
		currentLangTag := utils.LanguageTagFromWorldName(observation.Server)
		unique := true
		for _, langTag := range langTags {
			if langTag == currentLangTag {
				unique = false
				break
			}
		}
		if unique {
			langTags = append(langTags, currentLangTag)
		}
	}
	for server, tribeIDs := range tribeIDsByServer {
		list, err := s.cfg.API.Tribes.Browse(server, &shared_models.TribeFilter{
			ID: tribeIDs,
		})
		if err != nil {
			s.SendMessage(m.ChannelID, m.Author.Mention()+` Wystąpił błąd wewnętrzny, prosimy spróbować później.`)
			return
		}
		for _, tribe := range list.Items {
			observations[observationIndexByTribeID[tribe.ID]].Tribe = tribe
		}
	}
	langVersionList, err := s.cfg.API.LangVersions.Browse(&shared_models.LangVersionFilter{
		Tag: langTags,
	})

	msg := &EmbedMessage{}
	if len(observations) <= 0 {
		msg.Append("Brak")
	}
	for i, observation := range observations {
		tag := "Unknown"
		if observation.Tribe != nil {
			tag = observation.Tribe.Tag
		}
		lv := utils.FindLangVersionByTag(langVersionList.Items, utils.LanguageTagFromWorldName(observation.Server))
		tribeURL := ""
		if lv != nil {
			tribeURL = utils.FormatTribeURL(observation.Server, lv.Host, observation.TribeID)
		}
		msg.Append(fmt.Sprintf("**%d** | %d - %s - [``%s``](%s)\n", i+1, observation.ID,
			observation.Server,
			tag,
			tribeURL))
	}
	s.SendEmbed(m.ChannelID, NewEmbed().
		SetTitle("Lista obserwowanych plemion\nIndeks | ID - Serwer - Plemię").
		SetFields(msg.ToMessageEmbedFields()).
		SetFooter("Strona 1 z 1").
		MessageEmbed)
}
